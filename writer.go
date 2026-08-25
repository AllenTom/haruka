package haruka

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

// responseRecorder wraps an http.ResponseWriter to capture the status
// code and the number of bytes written, while transparently passing
// through Flush and Hijack so that SSE / WebSocket still work.
type responseRecorder struct {
	http.ResponseWriter
	status int
	size   int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w}
}

func (r *responseRecorder) WriteHeader(code int) {
	if r.status == 0 {
		r.status = code
	}
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(b)
	r.size += n
	return n, err
}

// Status returns the recorded status code, or 0 if nothing written yet.
func (r *responseRecorder) Status() int {
	return r.status
}

func (r *responseRecorder) Size() int {
	return r.size
}

// Written reports whether any header/body has been flushed.
func (r *responseRecorder) Written() bool {
	return r.status != 0
}

func (r *responseRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("haruka: underlying ResponseWriter does not support Hijack")
	}
	return h.Hijack()
}
