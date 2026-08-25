package middleware

import (
	"context"
	"time"

	"github.com/allentom/haruka"
)

// TimeoutMiddleware sets a deadline on the request context. Handlers
// and downstream code should honour ctx.Request.Context(); operations
// exceeding the deadline get cancelled with context.DeadlineExceeded.
//
// For hard enforcement against handlers that ignore the context, wrap
// the engine handler yourself (e.g. http.TimeoutHandler) or rely on
// Engine.ReadHeaderTimeout for header phases.
type TimeoutMiddleware struct {
	Duration time.Duration
}

func NewTimeoutMiddleware(duration time.Duration) *TimeoutMiddleware {
	return &TimeoutMiddleware{Duration: duration}
}

func (m *TimeoutMiddleware) OnRequest(ctx *haruka.Context) {
	if m.Duration <= 0 || ctx.Request == nil {
		return
	}
	timeoutCtx, cancel := context.WithTimeout(ctx.Request.Context(), m.Duration)
	// Intentionally not deferred: cancelling here would kill the
	// request immediately. Resources are reclaimed when the deadline
	// fires or the server closes the request.
	_ = cancel
	ctx.Request = ctx.Request.WithContext(timeoutCtx)
}
