package haruka

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

type Engine struct {
	Router      *Router
	Middlewares []Middleware
	server      *http.Server
	Logger      *logrus.Logger
	Cros        *cors.Cors
	// ReadHeaderTimeout guards against slowloris attacks.
	// Zero means the default of 15 seconds.
	ReadHeaderTimeout time.Duration
	// ShutdownTimeout is the maximum time to wait for in-flight
	// requests during graceful shutdown. Zero means 10 seconds.
	ShutdownTimeout time.Duration
}

func NewEngine() *Engine {
	return &Engine{
		Router:      NewRouter(),
		Middlewares: make([]Middleware, 0),
		Logger:      logrus.New(),
	}
}

// UseMiddleware registers a global middleware. It takes effect for
// requests handled after the call.
func (e *Engine) UseMiddleware(middleware Middleware) {
	e.Middlewares = append(e.Middlewares, middleware)
	e.Router.Middleware = e.Middlewares
}

func (e *Engine) UseCors(cors *cors.Cors) {
	e.Cros = cors
}

// HTTPHandler assembles the final root http.Handler (CORS included).
func (e *Engine) HTTPHandler() http.Handler {
	var router http.Handler = e.Router.HandlerRouter
	if e.Cros != nil {
		router = e.Cros.Handler(e.Router.HandlerRouter)
	}
	return router
}

func (e *Engine) buildServer(addr string, handler http.Handler) *http.Server {
	readHeaderTimeout := e.ReadHeaderTimeout
	if readHeaderTimeout == 0 {
		readHeaderTimeout = 15 * time.Second
	}
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
	}
}

func (e *Engine) RunAndListen(addr string) error {
	e.Router.Middleware = e.Middlewares
	e.Logger.Info(fmt.Sprintf("application run in %s", addr))
	httpserver := e.buildServer(addr, e.HTTPHandler())
	e.server = httpserver
	return httpserver.ListenAndServe()
}

// RunAndListenTLS serves HTTPS. certFile/keyFile are PEM paths.
func (e *Engine) RunAndListenTLS(addr, certFile, keyFile string) error {
	e.Router.Middleware = e.Middlewares
	e.Logger.Info(fmt.Sprintf("application run in %s (tls)", addr))
	httpserver := e.buildServer(addr, e.HTTPHandler())
	e.server = httpserver
	return httpserver.ListenAndServeTLS(certFile, keyFile)
}

// Shutdown gracefully drains in-flight connections. The context
// controls the deadline; callers should not wait forever.
func (e *Engine) Shutdown(ctx context.Context) error {
	if e.server == nil {
		return nil
	}
	return e.server.Shutdown(ctx)
}

// ShutdownWithTimeout gracefully drains in-flight connections, aborting
// after the given timeout.
func (e *Engine) ShutdownWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return e.Shutdown(ctx)
}

// Stop gracefully shuts the server down using ShutdownTimeout (default
// 10s). It will never block indefinitely.
func (e *Engine) Stop() error {
	timeout := e.ShutdownTimeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return e.ShutdownWithTimeout(timeout)
}
