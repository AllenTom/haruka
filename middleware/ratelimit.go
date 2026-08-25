package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/allentom/haruka"
	"golang.org/x/time/rate"
)

// clientEntry tracks one rate limiter per key.
type clientEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimitMiddleware applies a token bucket limiter.
//
// By default requests are keyed by client IP; provide KeyFunc to rate
// limit by API key, user id, etc. When the limit is hit the request is
// aborted with 429 and a Retry-After header.
type RateLimitMiddleware struct {
	// Rate is the sustained number of events per second.
	Rate rate.Limit
	// Burst is the maximum burst size.
	Burst int
	// KeyFunc derives the limiter key from the context.
	// Defaults to ClientIP.
	KeyFunc func(ctx *haruka.Context) string
	// Message returned in the JSON body on rejection.
	Message string

	mu       sync.Mutex
	clients  map[string]*clientEntry
	initOnce sync.Once
}

func NewRateLimitMiddleware(perSecond float64, burst int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		Rate:    rate.Limit(perSecond),
		Burst:   burst,
		Message: "rate limit exceeded",
	}
}

const rateLimitPruneThreshold = 4096
const rateLimitIdle = time.Hour

func (m *RateLimitMiddleware) keyOf(ctx *haruka.Context) string {
	if m.KeyFunc != nil {
		return m.KeyFunc(ctx)
	}
	return ctx.ClientIP()
}

func (m *RateLimitMiddleware) allow(key string) bool {
	m.initOnce.Do(func() {
		m.clients = make(map[string]*clientEntry)
	})
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.clients) > rateLimitPruneThreshold {
		now := time.Now()
		for k, entry := range m.clients {
			if now.Sub(entry.lastSeen) > rateLimitIdle {
				delete(m.clients, k)
			}
		}
	}
	entry, exists := m.clients[key]
	if !exists {
		entry = &clientEntry{limiter: rate.NewLimiter(m.Rate, m.Burst)}
		m.clients[key] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter.Allow()
}

func (m *RateLimitMiddleware) OnRequest(ctx *haruka.Context) {
	if !m.allow(m.keyOf(ctx)) {
		retryAfter := strconv.Itoa(maxInt(1, int(time.Second)/maxInt(1, int(m.Rate))))
		ctx.Writer.Header().Set("Retry-After", retryAfter)
		ctx.AbortWithStatusJSON(http.StatusTooManyRequests, haruka.JSON{
			"error": haruka.JSON{"code": http.StatusTooManyRequests, "message": m.Message},
		})
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
