package haruka

import (
	"net/http"
	"strconv"
)

type Context struct {
	Writer      http.ResponseWriter
	Request     *http.Request
	Parameters  map[string]string
	Param       map[string]interface{}
	isAbort     bool
	isInterrupt bool
}
type RequestHandler func(context *Context)

// get query string from url
func (c *Context) GetQueryString(key string) string {
	return c.Request.URL.Query().Get(key)
}

// get query string as []string
func (c *Context) GetQueryStrings(key string) []string {
	return c.Request.URL.Query()[key]
}

// get query string as int
func (c *Context) GetQueryInt(key string) (int, error) {
	rawValue := c.Request.URL.Query().Get(key)
	value, err := strconv.Atoi(rawValue)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// get path parameter as string
func (c *Context) GetPathParameterAsString(key string) string {
	return c.Parameters[key]
}

// get path parameter as int
func (c *Context) GetPathParameterAsInt(key string) (int, error) {
	rawValue := c.Parameters[key]
	value, err := strconv.Atoi(rawValue)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// interrupt middleware chain
func (c *Context) Abort() {
	c.isAbort = true
}

// interrupt middleware chain
func (c *Context) Interrupt() {
	c.isInterrupt = true
}
