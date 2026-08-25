package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/allentom/haruka"
)

// gzipResponseWriter transparently gzips the response body.
type gzipResponseWriter struct {
	http.ResponseWriter
	gz  *gzip.Writer
	mu  sync.Mutex
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	w.mu.Lock()
	if w.gz == nil {
		w.gz = gzip.NewWriter(w.ResponseWriter)
	}
	w.mu.Unlock()
	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	w.mu.Lock()
	if w.gz == nil {
		w.gz = gzip.NewWriter(w.ResponseWriter)
	}
	gz := w.gz
	w.mu.Unlock()
	return gz.Write(b)
}

// Flush pushes pending gzip bytes downstream so SSE keeps working.
func (w *gzipResponseWriter) Flush() {
	w.mu.Lock()
	if w.gz != nil {
		_ = w.gz.Flush()
	}
	w.mu.Unlock()
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// GzipMiddleware compresses responses for clients sending
// Accept-Encoding: gzip. Responses that already carry a
// Content-Encoding are left untouched.
type GzipMiddleware struct {
	// Level overrides the compression level; zero means DefaultCompression.
	Level int
}

func NewGzipMiddleware() *GzipMiddleware {
	return &GzipMiddleware{}
}

func acceptsGzip(ctx *haruka.Context) bool {
	for _, value := range ctx.Request.Header.Values("Accept-Encoding") {
		for _, part := range strings.Split(value, ",") {
			if strings.TrimSpace(strings.SplitN(part, ";", 2)[0]) == "gzip" {
				return true
			}
		}
	}
	return false
}

func (m *GzipMiddleware) OnRequest(ctx *haruka.Context) {
	if !acceptsGzip(ctx) {
		return
	}
	ctx.Writer.Header().Set("Content-Encoding", "gzip")
	ctx.Writer.Header().Add("Vary", "Accept-Encoding")
	ctx.Writer = &gzipResponseWriter{ResponseWriter: ctx.Writer}
}

// OnResponse finalizes the gzip stream after the handler finished.
func (m *GzipMiddleware) OnResponse(ctx *haruka.Context) {
	wrapper, ok := ctx.Writer.(*gzipResponseWriter)
	if !ok {
		return
	}
	wrapper.mu.Lock()
	defer wrapper.mu.Unlock()
	if wrapper.gz == nil {
		return
	}
	_ = wrapper.gz.Close()
	wrapper.gz = nil
}

var _ haruka.ResponseMiddleware = (*GzipMiddleware)(nil)
var _ io.Closer = (*gzip.Writer)(nil)

