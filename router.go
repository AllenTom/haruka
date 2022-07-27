package haruka

import (
	"github.com/gorilla/mux"
	"net/http"
)

type RouterMapping struct {
	Pattern     string
	HandlerFunc RequestHandler
}

type Router struct {
	Handlers      []RouterMapping
	HandlerRouter *mux.Router
	Middleware    []Middleware
}

func NewRouter() *Router {
	muxRouter := mux.NewRouter()
	return &Router{
		HandlerRouter: muxRouter,
		Handlers:      []RouterMapping{},
	}
}
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
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		ctx := r.MakeHandlerContext(writer, request, pattern)
		if ctx == nil {
			return
		}
		handler(ctx)
	})
}
func (r *Router) GET(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		ctx := r.MakeHandlerContext(writer, request, pattern)
		if ctx == nil {
			return
		}
		handler(ctx)
	}).Methods("GET")
}

func (r *Router) POST(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		ctx := r.MakeHandlerContext(writer, request, pattern)
		if ctx == nil {
			return
		}
		handler(ctx)
	}).Methods("POST")
}

func (r *Router) PUT(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		ctx := r.MakeHandlerContext(writer, request, pattern)
		if ctx == nil {
			return
		}
		handler(ctx)
	}).Methods("PUT")
}

func (r *Router) PATCH(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		ctx := r.MakeHandlerContext(writer, request, pattern)
		if ctx == nil {
			return
		}
		handler(ctx)
	}).Methods("PATCH")
}

func (r *Router) DELETE(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		ctx := r.MakeHandlerContext(writer, request, pattern)
		if ctx == nil {
			return
		}
		handler(ctx)
	}).Methods("DELETE")
}

func (r *Router) Static(pattern string, staticPath string) {
	r.HandlerRouter.PathPrefix(pattern).Handler(http.StripPrefix(pattern, http.FileServer(http.Dir(staticPath))))
}

func (r *Router) METHODS(pattern string, methods []string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		ctx := r.MakeHandlerContext(writer, request, pattern)
		if ctx == nil {
			return
		}
		handler(ctx)
	}).Methods(methods...)
}

func createContext(writer http.ResponseWriter, request *http.Request) *Context {
	return &Context{
		Writer:     writer,
		Request:    request,
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
