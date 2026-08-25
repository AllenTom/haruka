package haruka

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SSE (Server-Sent Events) helpers.
//
// Typical usage inside a handler:
//
//	err := ctx.InitSSE()
//	for {
//	    if err := ctx.SendSSEEvent("progress", map[string]int{"p": 50}); err != nil { break }
//	}

func (c *Context) sseFlusher() http.Flusher {
	f, ok := c.Writer.(http.Flusher)
	if !ok {
		return nil
	}
	return f
}

// InitSSE writes the event-stream headers and performs the first
// flush so browsers open the stream immediately.
func (c *Context) InitSSE() error {
	f := c.sseFlusher()
	if f == nil {
		return fmt.Errorf("haruka: response writer does not support flushing")
	}
	header := c.Writer.Header()
	if header.Get("Content-Type") == "" || header.Get("Content-Type") == "text/event-stream" {
		header.Set("Content-Type", "text/event-stream")
	}
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")
	header.Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	f.Flush()
	return nil
}

// SendSSEEvent writes a named event with JSON encoded data.
func (c *Context) SendSSEEvent(event string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.writeSSE(fmt.Sprintf("event: %s\ndata: %s\n\n", event, payload))
}

// SendSSEData writes a default (unnamed) event with JSON data.
func (c *Context) SendSSEData(data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.writeSSE(fmt.Sprintf("data: %s\n\n", payload))
}

func (c *Context) writeSSE(frame string) error {
	if _, err := fmt.Fprint(c.Writer, frame); err != nil {
		return err
	}
	if f := c.sseFlusher(); f != nil {
		f.Flush()
	}
	return nil
}
