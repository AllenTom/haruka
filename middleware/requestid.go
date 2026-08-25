package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/allentom/haruka"
)

const defaultRequestIDHeader = "X-Request-ID"

// RequestIDMiddleware assigns an identifier to every request.
//
// The id is read from the incoming header when present, otherwise
// generated. It is stored in ctx.Param["requestId"], written to the
// response header and injected into the request context so downstream
// code can fetch it via:
//
//	ctx.Request.Context().Value(haruka.RequestIDContextKey)
type RequestIDMiddleware struct {
	// HeaderName overrides the request/response header (default X-Request-ID).
	HeaderName string
	// Generator produces ids when the incoming header is empty.
	Generator func() string

	initOnce sync.Once
}

// RequestIDContextKey is the context key holding the request id.
type RequestIDContextKey struct{}

func (m *RequestIDMiddleware) header() string {
	if m.HeaderName != "" {
		return m.HeaderName
	}
	return defaultRequestIDHeader
}

func (m *RequestIDMiddleware) generator() func() string {
	m.initOnce.Do(func() {
		if m.Generator == nil {
			m.Generator = GenerateRequestID
		}
	})
	return m.Generator
}

// GenerateRequestID returns a random 32 character hex string.
func GenerateRequestID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	return hex.EncodeToString(buf)
}

func (m *RequestIDMiddleware) OnRequest(ctx *haruka.Context) {
	id := ctx.Request.Header.Get(m.header())
	if id == "" {
		id = m.generator()()
	}
	if id == "" {
		return
	}
	ctx.Param["requestId"] = id
	ctx.Writer.Header().Set(m.header(), id)
	if ctx.Request.Context() != nil {
		ctx.Request = ctx.Request.WithContext(
			context.WithValue(ctx.Request.Context(), RequestIDContextKey{}, id),
		)
	}
}
