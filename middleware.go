package haruka

// Middleware processes a request before it reaches the handler.
// Implementations may also implement ResponseMiddleware to receive
// a callback after the handler has finished.
type Middleware interface {
	OnRequest(ctx *Context)
}

// ResponseMiddleware is an optional interface that a Middleware can
// implement to observe (or finalize) the response after the handler
// returns. It is invoked even when the chain was aborted or the
// handler panicked.
type ResponseMiddleware interface {
	OnResponse(ctx *Context)
}
