package middleware

import (
	"net/http"

	"github.com/allentom/haruka"
)

// BodyLimitMiddleware caps the request body size using
// http.MaxBytesReader. Reads beyond the limit fail with an error
// ("http: request body too large"); handlers should map that to a
// 413 response.
type BodyLimitMiddleware struct {
	MaxBytes int64
}

func NewBodyLimitMiddleware(maxBytes int64) *BodyLimitMiddleware {
	return &BodyLimitMiddleware{MaxBytes: maxBytes}
}

func (m *BodyLimitMiddleware) OnRequest(ctx *haruka.Context) {
	if m.MaxBytes <= 0 || ctx.Request == nil || ctx.Request.Body == nil {
		return
	}
	// RawWriter is required so that net/http can turn the overflow
	// into an automatic 413 response after the handler returns.
	ctx.Request.Body = http.MaxBytesReader(ctx.RawWriter(), ctx.Request.Body, m.MaxBytes)
}
