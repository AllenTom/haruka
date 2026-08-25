package haruka

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gorilla/mux"
)

type RouterMapping struct {
	Pattern     string
	HandlerFunc RequestHandler
}

// RouteInfo describes a registered route, used for introspection.
type RouteInfo struct {
	Pattern string
	Methods []string
}

type routeEntry struct {
	pattern string
	methods []string
}

type Router struct {
	Handlers      []RouterMapping
	HandlerRouter *mux.Router
	Middleware    []Middleware
	routes        []routeEntry
}

func NewRouter() *Router {
	muxRouter := mux.NewRouter()
	return &Router{
		HandlerRouter: muxRouter,
		Handlers:      []RouterMapping{},
	}
}

// Routes returns a snapshot of all registered routes.
func (r *Router) Routes() []RouteInfo {
	out := make([]RouteInfo, 0, len(r.routes))
	for _, entry := range r.routes {
		out = append(out, RouteInfo{Pattern: entry.pattern, Methods: entry.methods})
	}
	return out
}

// Group creates a route group sharing the given prefix and middlewares.
func (r *Router) Group(prefix string, middlewares ...Middleware) *Group {
	return &Group{
		router:      r,
		prefix:      prefix,
		middlewares: append([]Middleware{}, middlewares...),
	}
}

func (r *Router) register(pattern string, methods []string, handler RequestHandler, chain []Middleware) {
	r.Handlers = append(r.Handlers, RouterMapping{Pattern: pattern, HandlerFunc: handler})
	r.routes = append(r.routes, routeEntry{pattern: pattern, methods: methods})
	route := r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		r.dispatch(writer, request, pattern, handler, chain)
	})
	if len(methods) > 0 {
		route.Methods(methods...)
	}
}

// MakeHandlerContext builds a Context and runs the global middleware
// chain. It returns nil when the chain was aborted. Kept for backward
// compatibility; the internal dispatch path uses more features.
func (r *Router) MakeHandlerContext(writer http.ResponseWriter, request *http.Request, pattern string) *Context {
	ctx := createContext(writer, request)
	ctx.Pattern = pattern
	r.execMiddleware(ctx)
	if ctx.isAbort {
		return nil
	}
	return ctx
}

func (r *Router) AddHandler(pattern string, handler RequestHandler) {
	r.register(pattern, nil, handler, nil)
}

func (r *Router) GET(pattern string, handler RequestHandler) {
	r.register(pattern, []string{"GET"}, handler, nil)
}

func (r *Router) POST(pattern string, handler RequestHandler) {
	r.register(pattern, []string{"POST"}, handler, nil)
}

func (r *Router) PUT(pattern string, handler RequestHandler) {
	r.register(pattern, []string{"PUT"}, handler, nil)
}

func (r *Router) PATCH(pattern string, handler RequestHandler) {
	r.register(pattern, []string{"PATCH"}, handler, nil)
}

func (r *Router) DELETE(pattern string, handler RequestHandler) {
	r.register(pattern, []string{"DELETE"}, handler, nil)
}

func (r *Router) Static(pattern string, staticPath string) {
	r.routes = append(r.routes, routeEntry{pattern: pattern + "/", methods: []string{"GET"}})
	r.HandlerRouter.PathPrefix(pattern).Handler(http.StripPrefix(pattern, http.FileServer(http.Dir(staticPath))))
}

func (r *Router) METHODS(pattern string, methods []string, handler RequestHandler) {
	r.register(pattern, methods, handler, nil)
}

func (r *Router) Prefix(pattern string, handler RequestHandler) {
	r.routes = append(r.routes, routeEntry{pattern: pattern + "/*", methods: nil})
	r.HandlerRouter.PathPrefix(pattern).HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		r.dispatch(writer, request, pattern, handler, nil)
	})
}

func createContext(writer http.ResponseWriter, request *http.Request) *Context {
	recorder := newResponseRecorder(writer)
	return &Context{
		Writer:     recorder,
		recorder:   recorder,
		Request:    request,
		startTime:  time.Now(),
		Parameters: mux.Vars(request),
		Param:      map[string]interface{}{},
		isAbort:    false,
	}
}

func (r *Router) execMiddleware(ctx *Context) {
	for _, middleware := range r.Middleware {
		middleware.OnRequest(ctx)
		if ctx.isAbort {
			return
		}
	}
}

// dispatch is the central request pipeline:
//
//	global middlewares -> group middlewares -> handler
//
// with panic recovery and ResponseMiddleware hooks around it.
func (r *Router) dispatch(writer http.ResponseWriter, request *http.Request, pattern string, handler RequestHandler, chain []Middleware) {
	ctx := createContext(writer, request)
	ctx.Pattern = pattern

	combined := make([]Middleware, 0, len(r.Middleware)+len(chain))
	combined = append(combined, r.Middleware...)
	combined = append(combined, chain...)

	defer func() {
		if rec := recover(); rec != nil {
			recoverFromPanic(ctx, rec)
		}
		for _, mw := range combined {
			if rm, ok := mw.(ResponseMiddleware); ok {
				r.safeOnResponse(rm, ctx)
			}
		}
	}()

	for _, mw := range combined {
		mw.OnRequest(ctx)
		if ctx.isAbort {
			return
		}
	}
	handler(ctx)
}

func (r *Router) safeOnResponse(rm ResponseMiddleware, ctx *Context) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("[haruka] panic in OnResponse: %v\n%s", rec, debug.Stack())
		}
	}()
	rm.OnResponse(ctx)
}

var recoveryHandler = defaultRecoveryHandler

// SetRecoveryHandler replaces the built-in panic recovery handler.
// The default one logs the stack and renders a structured 500 error
// (or the panicked *HTTPError itself).
func SetRecoveryHandler(handler func(ctx *Context, panicValue interface{})) {
	if handler != nil {
		recoveryHandler = handler
	} else {
		recoveryHandler = defaultRecoveryHandler
	}
}

func recoverFromPanic(ctx *Context, panicValue interface{}) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("[haruka] panic inside recovery handler: %v\n%s", rec, debug.Stack())
		}
	}()
	recoveryHandler(ctx, panicValue)
}

func defaultRecoveryHandler(ctx *Context, panicValue interface{}) {
	log.Printf("[haruka] panic recovered on %s %s: %v\n%s",
		ctx.Request.Method, ctx.Request.URL.Path, panicValue, debug.Stack())
	httpErr, ok := panicValue.(*HTTPError)
	if !ok {
		httpErr = ErrInternal("Internal Server Error")
	}
	if !ctx.recorder.Written() {
		ctx.AbortWithError(httpErr)
	}
}
