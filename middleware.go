package haruka

type Middleware interface {
	OnRequest(ctx *Context)
}
