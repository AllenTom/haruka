# haruka
web framework

## Example

```go
package main

import "github.com/allentom/haruka"

func main() {
	e := haruka.NewEngine()
	e.Router.GET("/ping", func(context *Context) {
		context.JSONWithStatus(haruka.JSON{
			"message":"Pong",
		}, http.StatusOK)
    })
	e.ListenAndServe(":8080")
}
```
## Router
### With specific path

```go
e.Router.GET("/ping", <YourHandler>)
// with path parameters
e.Router.GET("/ping/{name}", <YourHandler>)
// with regex
e.Router.GET("/ping/{name:[0-9|a-z|A-Z]+}", <YourHandler>)
```
### With different methods

```go
e.Router.GET("/ping", <YourHandler>)
e.Router.POST("/ping", <YourHandler>)
e.Router.PATCH("/ping", <YourHandler>)
e.Router.Put("/ping", <YourHandler>)
e.Router.OPTION("/ping", <YourHandler>)

// for all methods
e.Router.AddHandler("/ping", <YourHandler>)
// for multiple methods
e.Router.METHODS("/ping", []string{"GET","POST"}, <YourHandler>)
```

## Handler

### Get query parameters

```go
func (c *Context) {
    context.GetQueryString("id")
    context.GetQueryStrings("ids")
    context.GetQueryInt("id")
}
```

### Get path parameters

```go
func (c *Context) {
    context.GetPathParameterAsInt("id")
    context.GetPathParameterAsString("id")
}

```

## Router Groups

Routes can share a prefix and a middleware chain. Groups nest; the
effective chain is outermost-first.

```go
api := e.Router.Group("/api", authMiddleware)
api.GET("/users", listUsers)

admin := api.Group("/admin", adminOnly)
admin.GET("/stats", stats) // /api/admin/stats
```

## Middleware

A middleware implements `OnRequest(*Context)`. It may also implement
`OnResponse(*Context)` (interface `ResponseMiddleware`) to observe the
response after the handler returns — invoked even when aborted or
after a recovered panic.

Built-in middlewares (`github.com/allentom/haruka/middleware`):

| Middleware | Description |
|------------|-------------|
| `NewLoggerMiddleware` | Logs method, path, status, duration, ip and request id on response |
| `&RequestIDMiddleware{}` | Assigns/propagates `X-Request-ID`, stores it in `ctx.Param["requestId"]` |
| `NewTimeoutMiddleware(d)` | Sets a context deadline for the request |
| `NewRateLimitMiddleware(rps, burst)` | Token bucket rate limiting keyed by client IP |
| `NewGzipMiddleware()` | Transparent gzip response compression |
| `NewBodyLimitMiddleware(n)` | Caps request body size |
| `&CSRFMiddleware{}` | Double-submit cookie CSRF protection |
| `NewJWTMiddleware` | JWT parsing (now powered by `golang-jwt/jwt/v5`) |

## Panic Recovery

Panics in handlers/middlewares are always recovered: the default
handler logs the stack and renders a structured 500 error. A panic
carrying an `*HTTPError` is rendered with its own status and message.
Customize with:

```go
haruka.SetRecoveryHandler(func(ctx *haruka.Context, panicValue interface{}) {
	// your rendering
})
```

## Structured Errors

```go
err := haruka.ErrNotFound("item vanished") // 404 + {"error":{"code":404,"message":"..."}}
ctx.JSONError(err)        // write envelope
ctx.AbortWithError(err)   // write envelope and abort chain

// or simply:
panic(haruka.ErrForbidden("no access"))
```

## Server Lifecycle

```go
e.RunAndListen(":8080")
e.RunAndListenTLS(":443", "cert.pem", "key.pem")
e.Shutdown(ctx)   // graceful drain
e.Stop()          // graceful shutdown, waits indefinitely
```

`Engine.ReadHeaderTimeout` guards slowloris-style connections
(default 15s).

## SSE (Server-Sent Events)

```go
func handler(c *haruka.Context) {
	c.InitSSE()
	c.SendSSEEvent("progress", map[string]int{"p": 50})
	c.SendSSEData("done")
}
```

## WebSocket

```go
func handler(c *haruka.Context) {
	conn, err := c.UpgradeToWebSocket()
	// github.com/gorilla/websocket conn
}
```

## Context Helpers

```go
c.StatusCode()   // status written so far (0 = nothing yet)
c.ClientIP()     // X-Forwarded-For > X-Real-IP > RemoteAddr
c.StartTime()    // when the request entered the framework
e.Router.Routes() // registered route listing
```
