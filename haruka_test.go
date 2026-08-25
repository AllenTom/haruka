package haruka

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// newTestServer serves the engine over an httptest server and closes
// it automatically when the test ends.
func newTestServer(t *testing.T, e *Engine) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(e.HTTPHandler())
	t.Cleanup(server.Close)
	return server
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	return string(body)
}

func doGet(t *testing.T, url string) (*http.Response, string) {
	t.Helper()
	resp, err := http.Get(url)
	assert.NoError(t, err)
	defer resp.Body.Close()
	return resp, readBody(t, resp)
}

func doPost(t *testing.T, url string) (*http.Response, string) {
	t.Helper()
	resp, err := http.Post(url, "application/json", nil)
	assert.NoError(t, err)
	defer resp.Body.Close()
	return resp, readBody(t, resp)
}

func doPut(t *testing.T, url string) (*http.Response, string) {
	return doMethod(t, http.MethodPut, url)
}

func doPatch(t *testing.T, url string) (*http.Response, string) {
	return doMethod(t, http.MethodPatch, url)
}

func doDelete(t *testing.T, url string) (*http.Response, string) {
	return doMethod(t, http.MethodDelete, url)
}

func doMethod(t *testing.T, method string, url string) (*http.Response, string) {
	t.Helper()
	request, err := http.NewRequest(method, url, nil)
	assert.NoError(t, err)
	resp, err := http.DefaultClient.Do(request)
	assert.NoError(t, err)
	defer resp.Body.Close()
	return resp, readBody(t, resp)
}

func TestRunAndListen(t *testing.T) {
	e := NewEngine()
	e.Router.AddHandler("/ping", func(context *Context) {
		err := context.JSON(JSON{"say": "pong"})
		if err != nil {
			t.Error(err)
		}
	})
	server := newTestServer(t, e)
	resp, body := doGet(t, server.URL+"/ping")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "{\"say\":\"pong\"}", body)
}

type XMLBody struct {
	XMLName xml.Name `xml:"test"`
	Say     string   `xml:"say"`
}

func TestXML(t *testing.T) {
	e := NewEngine()
	e.Router.AddHandler("/ping", func(context *Context) {
		err := context.XML(XMLBody{Say: "Hello"})
		if err != nil {
			t.Error(err)
		}
	})
	server := newTestServer(t, e)
	resp, body := doGet(t, server.URL+"/ping")
	assert.Equal(t, "200 OK", resp.Status)
	assert.Equal(t, "<test><say>Hello</say></test>", body)
}

func TestRouter_GET(t *testing.T) {
	e := NewEngine()
	e.Router.GET("/ping", func(context *Context) {
		err := context.JSON(JSON{"say": "pong"})
		if err != nil {
			t.Error(err)
		}
	})
	server := newTestServer(t, e)
	resp, _ := doPost(t, server.URL+"/ping")
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	getResp, getBody := doGet(t, server.URL+"/ping")
	assert.Equal(t, "{\"say\":\"pong\"}", getBody)
	assert.Equal(t, http.StatusOK, getResp.StatusCode)
}

func TestRouter_POST(t *testing.T) {
	e := NewEngine()
	e.Router.POST("/ping", func(context *Context) {
		err := context.JSON(JSON{"say": "pong"})
		if err != nil {
			t.Error(err)
		}
	})
	server := newTestServer(t, e)
	resp, body := doPost(t, server.URL+"/ping")
	assert.Equal(t, "{\"say\":\"pong\"}", body)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// other methods 405 response
	methodResp, _ := doPatch(t, server.URL+"/ping")
	assert.Equal(t, http.StatusMethodNotAllowed, methodResp.StatusCode)
}

func TestRouter_PUT(t *testing.T) {
	e := NewEngine()
	e.Router.PUT("/ping", func(context *Context) {
		err := context.JSON(JSON{"say": "pong"})
		if err != nil {
			t.Error(err)
		}
	})
	server := newTestServer(t, e)
	_, body := doPut(t, server.URL+"/ping")
	assert.Equal(t, "{\"say\":\"pong\"}", body)
	// other methods 405 response
	postResp, _ := doPost(t, server.URL+"/ping")
	assert.Equal(t, http.StatusMethodNotAllowed, postResp.StatusCode)
}

func TestRouter_PATCH(t *testing.T) {
	e := NewEngine()
	e.Router.PATCH("/ping", func(context *Context) {
		err := context.JSON(JSON{"say": "pong"})
		if err != nil {
			t.Error(err)
		}
	})
	server := newTestServer(t, e)
	_, body := doPatch(t, server.URL+"/ping")
	assert.Equal(t, "{\"say\":\"pong\"}", body)
	// other methods 405 response
	postResp, _ := doPost(t, server.URL+"/ping")
	assert.Equal(t, http.StatusMethodNotAllowed, postResp.StatusCode)
}

func TestRouter_DELETE(t *testing.T) {
	e := NewEngine()
	e.Router.DELETE("/ping", func(context *Context) {
		err := context.JSON(JSON{"say": "pong"})
		if err != nil {
			t.Error(err)
		}
	})
	server := newTestServer(t, e)
	_, body := doDelete(t, server.URL+"/ping")
	assert.Equal(t, "{\"say\":\"pong\"}", body)
	// other methods 405 response
	postResp, _ := doPost(t, server.URL+"/ping")
	assert.Equal(t, http.StatusMethodNotAllowed, postResp.StatusCode)
}

func TestRouter_Prefix(t *testing.T) {
	e := NewEngine()
	e.Router.Prefix("/api", func(context *Context) {
		data := map[string]interface{}{
			"path": context.Request.URL.Path,
		}
		err := context.JSON(data)
		if err != nil {
			t.Error(err)
		}
	})
	server := newTestServer(t, e)

	// Test exact prefix match
	_, body := doGet(t, server.URL+"/api")
	assert.Equal(t, "{\"path\":\"/api\"}", body)

	// Test subpath match
	_, body = doGet(t, server.URL+"/api/users")
	assert.Equal(t, "{\"path\":\"/api/users\"}", body)

	// Test deeper subpath match
	_, body = doGet(t, server.URL+"/api/users/123/profile")
	assert.Equal(t, "{\"path\":\"/api/users/123/profile\"}", body)

	// Test non-matching path should return 404
	resp, _ := doGet(t, server.URL+"/not-api")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
