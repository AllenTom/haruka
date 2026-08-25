package haruka_test

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	haruka "github.com/allentom/haruka"
	"github.com/allentom/haruka/middleware"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func logrusNewSilent() *logrus.Logger {
	l := logrus.New()
	l.Out = io.Discard
	return l
}

type memoryHook struct {
	mu      sync.Mutex
	entries []*logrus.Entry
}

func (h *memoryHook) Fire(entry *logrus.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = append(h.entries, entry)
	return nil
}

func (h *memoryHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *memoryHook) lastEntry() *logrus.Entry {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.entries) == 0 {
		return nil
	}
	return h.entries[len(h.entries)-1]
}

// newTestServer serves the engine over an httptest server and closes
// it automatically when the test ends.
func newTestServer(t *testing.T, e *haruka.Engine) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(e.HTTPHandler())
	t.Cleanup(server.Close)
	return server
}

func doGet(t *testing.T, url string) (*http.Response, string) {
	t.Helper()
	resp, err := http.Get(url)
	assert.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	return resp, string(body)
}

func doPost(t *testing.T, url string) (*http.Response, string) {
	t.Helper()
	resp, err := http.Post(url, "application/json", nil)
	assert.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	return resp, string(body)
}

// ---------- Router Group ----------

type fakeMiddleware struct {
	onRequest func(ctx *haruka.Context)
}

func (f *fakeMiddleware) OnRequest(ctx *haruka.Context) {
	f.onRequest(ctx)
}

func TestGroup_PrefixAndMiddleware(t *testing.T) {
	e := haruka.NewEngine()
	var publicCalled, groupCalled, adminCalled int32

	groupAuth := &fakeMiddleware{
		onRequest: func(ctx *haruka.Context) {
			atomic.AddInt32(&groupCalled, 1)
			if ctx.GetQueryString("token") != "secret" {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, haruka.JSON{"error": "no token"})
			}
		},
	}

	e.Router.GET("/public", func(ctx *haruka.Context) {
		atomic.AddInt32(&publicCalled, 1)
		ctx.JSON(haruka.JSON{"ok": true})
	})

	api := e.Router.Group("/api", groupAuth)
	api.GET("/users", func(ctx *haruka.Context) {
		ctx.JSON(haruka.JSON{"users": []string{"a"}})
	})

	admin := api.Group("/admin")
	admin.GET("/stats", func(ctx *haruka.Context) {
		atomic.AddInt32(&adminCalled, 1)
		ctx.JSON(haruka.JSON{"stats": true})
	})

	server := newTestServer(t, e)

	resp, body := doGet(t, server.URL+"/public")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "{\"ok\":true}", body)
	assert.Equal(t, int32(1), atomic.LoadInt32(&publicCalled))

	// blocked by group middleware
	resp, body = doGet(t, server.URL+"/api/users")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.JSONEq(t, "{\"error\":\"no token\"}", body)

	// passes with token
	resp, _ = doGet(t, server.URL+"/api/users?token=secret")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(2), atomic.LoadInt32(&groupCalled))

	// nested group inherits parent middleware
	resp, _ = doGet(t, server.URL+"/api/admin/stats?token=secret")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&adminCalled))
	assert.Equal(t, int32(3), atomic.LoadInt32(&groupCalled))

	var patterns []string
	for _, r := range e.Router.Routes() {
		patterns = append(patterns, r.Pattern)
	}
	assert.Contains(t, patterns, "/api/users")
	assert.Contains(t, patterns, "/api/admin/stats")
}

func TestGroup_UseMiddlewareAfterCreation(t *testing.T) {
	e := haruka.NewEngine()
	g := e.Router.Group("/g")
	called := false
	g.UseMiddleware(&fakeMiddleware{
		onRequest: func(ctx *haruka.Context) { called = true },
	})
	g.GET("/ping", func(ctx *haruka.Context) { ctx.JSON(haruka.JSON{"ok": true}) })
	server := newTestServer(t, e)
	doGet(t, server.URL+"/g/ping")
	assert.True(t, called)
}

// ---------- Recovery ----------

func TestRecovery_DefaultPanic(t *testing.T) {
	e := haruka.NewEngine()
	e.Router.GET("/boom", func(ctx *haruka.Context) {
		panic("explode")
	})
	server := newTestServer(t, e)
	resp, body := doGet(t, server.URL+"/boom")
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	var envelope struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	assert.NoError(t, json.Unmarshal([]byte(body), &envelope))
	assert.Equal(t, "Internal Server Error", envelope.Error.Message)
}

func TestRecovery_HTTPErrorPanic(t *testing.T) {
	e := haruka.NewEngine()
	e.Router.GET("/gone", func(ctx *haruka.Context) {
		panic(haruka.ErrNotFound("item vanished"))
	})
	server := newTestServer(t, e)
	resp, body := doGet(t, server.URL+"/gone")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.JSONEq(t, "{\"error\":{\"code\":404,\"message\":\"item vanished\"}}", body)
}

func TestJSONError_AbortWithError(t *testing.T) {
	e := haruka.NewEngine()
	e.Router.POST("/login", func(ctx *haruka.Context) {
		ctx.AbortWithError(haruka.ErrUnauthorized("bad credentials"))
	})
	server := newTestServer(t, e)
	resp, body := doPost(t, server.URL+"/login")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.JSONEq(t, "{\"error\":{\"code\":401,\"message\":\"bad credentials\"}}", body)
}

// ---------- RequestID ----------

func TestRequestIDMiddleware(t *testing.T) {
	e := haruka.NewEngine()
	e.UseMiddleware(&middleware.RequestIDMiddleware{})
	var seenID string
	e.Router.GET("/id", func(ctx *haruka.Context) {
		seenID, _ = ctx.Param["requestId"].(string)
		ctx.JSON(haruka.JSON{"ok": true})
	})
	server := newTestServer(t, e)
	resp, _ := doGet(t, server.URL+"/id")
	headerID := resp.Header.Get("X-Request-ID")
	assert.NotEmpty(t, headerID)
	assert.Equal(t, headerID, seenID)

	req, _ := http.NewRequest("GET", server.URL+"/id", nil)
	req.Header.Set("X-Request-ID", "my-custom-id")
	resp2, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, "my-custom-id", resp2.Header.Get("X-Request-ID"))
}

// ---------- Timeout ----------

func TestTimeoutMiddleware_DeadlineVisible(t *testing.T) {
	e := haruka.NewEngine()
	e.UseMiddleware(middleware.NewTimeoutMiddleware(50 * time.Millisecond))
	e.Router.GET("/slow", func(ctx *haruka.Context) {
		select {
		case <-ctx.Request.Context().Done():
			_ = ctx.JSONError(haruka.ErrInternal("deadline exceeded"))
		case <-time.After(2 * time.Second):
			ctx.JSON(haruka.JSON{"ok": true})
		}
	})
	server := newTestServer(t, e)
	start := time.Now()
	resp, body := doGet(t, server.URL+"/slow")
	assert.Less(t, int64(time.Since(start)), int64(time.Second))
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Contains(t, body, "deadline exceeded")
}

// ---------- BodyLimit ----------

func TestBodyLimitMiddleware_RejectsOversize(t *testing.T) {
	e := haruka.NewEngine()
	e.UseMiddleware(middleware.NewBodyLimitMiddleware(16)) // 16 bytes
	e.Router.POST("/upload", func(ctx *haruka.Context) {
		_, err := io.Copy(io.Discard, ctx.Request.Body)
		if err != nil {
			ctx.AbortWithError(haruka.ErrBadRequest("request body too large"))
			return
		}
		ctx.JSON(haruka.JSON{"ok": true})
	})
	server := newTestServer(t, e)
	resp, err := http.Post(server.URL+"/upload", "text/plain", strings.NewReader(strings.Repeat("x", 1024)))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	smallResp, err := http.Post(server.URL+"/upload", "text/plain", strings.NewReader("tiny"))
	assert.NoError(t, err)
	defer smallResp.Body.Close()
	assert.Equal(t, http.StatusOK, smallResp.StatusCode)
}

// ---------- RateLimit ----------

func TestRateLimitMiddleware_BurstExceeded(t *testing.T) {
	e := haruka.NewEngine()
	e.UseMiddleware(middleware.NewRateLimitMiddleware(0.0001, 2))
	e.Router.GET("/limited", func(ctx *haruka.Context) { ctx.JSON(haruka.JSON{"ok": true}) })
	server := newTestServer(t, e)

	for i := 0; i < 2; i++ {
		resp, _ := doGet(t, server.URL+"/limited")
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
	resp, body := doGet(t, server.URL+"/limited")
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("Retry-After"))
	assert.Contains(t, body, "rate limit exceeded")
}

// ---------- CSRF ----------

func csrfTestClient(t *testing.T) *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar}
}

func TestCSRFMiddleware(t *testing.T) {
	e := haruka.NewEngine()
	csrf := &middleware.CSRFMiddleware{Secure: boolPtr(false)}
	e.UseMiddleware(csrf)
	var tokenFromCtx string
	e.Router.GET("/form", func(ctx *haruka.Context) {
		tokenFromCtx, _ = ctx.Param["csrfToken"].(string)
		ctx.JSON(haruka.JSON{"ok": true})
	})
	e.Router.POST("/submit", func(ctx *haruka.Context) {
		ctx.JSON(haruka.JSON{"ok": true})
	})
	server := newTestServer(t, e)
	client := csrfTestClient(t)

	// GET issues the cookie and exposes token to handler
	getReq, _ := http.NewRequest("GET", server.URL+"/form", nil)
	getResp, err := client.Do(getReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	_, err = io.ReadAll(getResp.Body)
	assert.NoError(t, err)
	getResp.Body.Close()
	assert.NotEmpty(t, tokenFromCtx)

	// POST without header -> 403
	req, _ := http.NewRequest("POST", server.URL+"/submit", nil)
	blocked, err := client.Do(req)
	assert.NoError(t, err)
	blockedBody, _ := io.ReadAll(blocked.Body)
	blocked.Body.Close()
	assert.Equal(t, http.StatusForbidden, blocked.StatusCode)
	assert.Contains(t, string(blockedBody), "CSRF")

	// POST with matching header -> 200
	req2, _ := http.NewRequest("POST", server.URL+"/submit", nil)
	req2.Header.Set("X-CSRF-Token", tokenFromCtx)
	ok, err := client.Do(req2)
	assert.NoError(t, err)
	okBody, _ := io.ReadAll(ok.Body)
	ok.Body.Close()
	assert.Equal(t, http.StatusOK, ok.StatusCode)
	assert.Contains(t, string(okBody), "ok")

	// POST with wrong header -> 403
	req3, _ := http.NewRequest("POST", server.URL+"/submit", nil)
	req3.Header.Set("X-CSRF-Token", "forged")
	bad, err := client.Do(req3)
	assert.NoError(t, err)
	bad.Body.Close()
	assert.Equal(t, http.StatusForbidden, bad.StatusCode)
}

func boolPtr(v bool) *bool { return &v }

// ---------- Gzip ----------

func TestGzipMiddleware_CompressesResponse(t *testing.T) {
	e := haruka.NewEngine()
	e.UseMiddleware(middleware.NewGzipMiddleware())
	payload := []byte("[")
	for i := 0; i < 64; i++ {
		if i > 0 {
			payload = append(payload, ',')
		}
		payload = append(payload, `{"hello":"world"}`...)
	}
	payload = append(payload, ']')
	e.Router.GET("/data", func(ctx *haruka.Context) { ctx.JSON(json.RawMessage(payload)) })
	server := newTestServer(t, e)

	req, _ := http.NewRequest("GET", server.URL+"/data", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	client := &http.Client{Transport: &http.Transport{DisableCompression: true}}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "gzip", resp.Header.Get("Content-Encoding"))
	raw, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	gr, err := gzip.NewReader(bytes.NewReader(raw))
	assert.NoError(t, err)
	decompressed, err := io.ReadAll(gr)
	assert.NoError(t, err)
	assert.Equal(t, string(payload), string(decompressed))
}

func TestGzipMiddleware_PlainWhenNotAccepted(t *testing.T) {
	e := haruka.NewEngine()
	e.UseMiddleware(middleware.NewGzipMiddleware())
	e.Router.GET("/data", func(ctx *haruka.Context) { ctx.JSON(haruka.JSON{"hello": "world"}) })
	server := newTestServer(t, e)

	resp, body := doGet(t, server.URL+"/data")
	assert.Empty(t, resp.Header.Get("Content-Encoding"))
	assert.JSONEq(t, "{\"hello\":\"world\"}", body)
}

// ---------- SSE ----------

func TestSSE_EventStream(t *testing.T) {
	e := haruka.NewEngine()
	e.Router.GET("/events", func(ctx *haruka.Context) {
		if err := ctx.InitSSE(); err != nil {
			t.Error(err)
			return
		}
		if err := ctx.SendSSEEvent("progress", map[string]int{"p": 50}); err != nil {
			t.Error(err)
		}
		if err := ctx.SendSSEData(map[string]string{"done": "yes"}); err != nil {
			t.Error(err)
		}
	})
	server := newTestServer(t, e)
	resp, err := http.Get(server.URL + "/events")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	expected := "event: progress\ndata: {\"p\":50}\n\ndata: {\"done\":\"yes\"}\n\n"
	assert.Equal(t, expected, string(body))
}

// ---------- WebSocket ----------

func TestWebSocket_Echo(t *testing.T) {
	e := haruka.NewEngine()
	e.Router.GET("/ws", func(ctx *haruka.Context) {
		conn, err := ctx.UpgradeToWebSocket()
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			mt, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if err := conn.WriteMessage(mt, msg); err != nil {
				return
			}
		}
	})
	server := newTestServer(t, e)
	url := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	assert.NoError(t, err)
	defer conn.Close()

	assert.NoError(t, conn.WriteMessage(websocket.TextMessage, []byte("hello haruka")))
	_, msg, err := conn.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, "hello haruka", string(msg))
}

// ---------- StatusCode helper / ResponseMiddleware ----------

func TestContext_StatusCodeHelper(t *testing.T) {
	e := haruka.NewEngine()
	logger := &statusProbe{}
	e.UseMiddleware(logger)
	e.Router.GET("/ok", func(ctx *haruka.Context) {
		assert.Equal(t, 0, ctx.StatusCode()) // nothing written yet
		ctx.JSONWithStatus(haruka.JSON{"ok": true}, http.StatusAccepted)
	})
	server := newTestServer(t, e)
	resp, _ := doGet(t, server.URL+"/ok")
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, int32(http.StatusAccepted), atomic.LoadInt32(&logger.lastStatus))
}

type statusProbe struct {
	lastStatus int32
}

func (p *statusProbe) OnRequest(ctx *haruka.Context) {}

func (p *statusProbe) OnResponse(ctx *haruka.Context) {
	atomic.StoreInt32(&p.lastStatus, int32(ctx.StatusCode()))
}

var _ haruka.ResponseMiddleware = (*statusProbe)(nil)

// ---------- Logger OnResponse hook ----------

func TestLoggerMiddleware_LogsStatusAndDuration(t *testing.T) {
	hook := &memoryHook{}
	silent := logrusNewSilent()
	silent.AddHook(hook)

	e := haruka.NewEngine()
	e.UseMiddleware(&middleware.RequestLoggerMiddleware{Logger: silent})
	e.Router.GET("/logged", func(ctx *haruka.Context) {
		time.Sleep(5 * time.Millisecond)
		ctx.JSON(haruka.JSON{"ok": true})
	})
	server := newTestServer(t, e)
	resp, _ := doGet(t, server.URL+"/logged")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	entry := hook.lastEntry()
	if assert.NotNil(t, entry) {
		assert.Equal(t, http.StatusOK, entry.Data["status"])
		assert.Equal(t, "/logged", entry.Data["path"])
		_, hasDuration := entry.Data["duration"]
		assert.True(t, hasDuration)
	}
}
