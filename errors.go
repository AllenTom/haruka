package haruka

import (
	"fmt"
	"net/http"
)

// HTTPError is a structured API error. Handlers may return it to the
// client via JSONError / AbortWithError, or simply panic with it —
// the built-in recovery renders it with its status code.
type HTTPError struct {
	// StatusCode is the HTTP status written to the response.
	// It is not serialized into the body.
	StatusCode int `json:"-"`
	// Code is an application specific error code (optional).
	Code int `json:"code,omitempty"`
	// Message is a human readable description.
	Message string `json:"message"`
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http %d: %s", e.StatusCode, e.Message)
}

func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{StatusCode: statusCode, Message: message}
}

func NewHTTPErrorWithCode(statusCode, code int, message string) *HTTPError {
	return &HTTPError{StatusCode: statusCode, Code: code, Message: message}
}

func ErrBadRequest(message string) *HTTPError {
	return NewHTTPErrorWithCode(http.StatusBadRequest, http.StatusBadRequest, message)
}

func ErrUnauthorized(message string) *HTTPError {
	return NewHTTPErrorWithCode(http.StatusUnauthorized, http.StatusUnauthorized, message)
}

func ErrForbidden(message string) *HTTPError {
	return NewHTTPErrorWithCode(http.StatusForbidden, http.StatusForbidden, message)
}

func ErrNotFound(message string) *HTTPError {
	return NewHTTPErrorWithCode(http.StatusNotFound, http.StatusNotFound, message)
}

func ErrInternal(message string) *HTTPError {
	return NewHTTPErrorWithCode(http.StatusInternalServerError, http.StatusInternalServerError, message)
}

type errorEnvelope struct {
	Error *HTTPError `json:"error"`
}

// JSONError writes err as a structured JSON envelope:
//
//	{"error":{"code":400,"message":"..."}}
func (c *Context) JSONError(err *HTTPError) error {
	return c.JSONWithStatus(errorEnvelope{Error: err}, err.StatusCode)
}

// AbortWithError writes err via JSONError and aborts the chain.
func (c *Context) AbortWithError(err *HTTPError) {
	_ = c.JSONError(err)
	c.Abort()
}
