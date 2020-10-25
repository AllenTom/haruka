package Haruka

import "net/http"

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
}
type RequestHandler func(context *Context)
