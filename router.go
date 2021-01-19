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
func (r *Router) AddHandler(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		ctx := createContext(writer, request)
		r.execMiddleware(ctx)
		handler(ctx)
	})
}

func (r *Router) GET(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		ctx := createContext(writer, request)
		handler(ctx)
		r.execMiddleware(ctx)
	}).Methods("GET")
}

func (r *Router) POST(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		ctx := createContext(writer, request)
		handler(ctx)
		r.execMiddleware(ctx)
	}).Methods("POST")
}

func (r *Router) PUT(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		ctx := createContext(writer, request)
		handler(ctx)
		r.execMiddleware(ctx)
	}).Methods("PUT")
}

func (r *Router) PATCH(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		ctx := createContext(writer, request)
		handler(ctx)
		r.execMiddleware(ctx)
	}).Methods("PATCH")
}

func (r *Router) DELETE(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		ctx := createContext(writer, request)
		handler(ctx)
		r.execMiddleware(ctx)
	}).Methods("DELETE")
}

func (r *Router) Static(pattern string, staticPath string) {
	r.HandlerRouter.PathPrefix(pattern).Handler(http.StripPrefix(pattern, http.FileServer(http.Dir(staticPath))))
}

func (r *Router) METHODS(pattern string, methods []string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		ctx := createContext(writer, request)
		handler(ctx)
		r.execMiddleware(ctx)
	}).Methods(methods...)
}

func createContext(writer http.ResponseWriter, request *http.Request) *Context {
	return &Context{
		Writer:     writer,
		Request:    request,
		Parameters: mux.Vars(request),
		Param:      map[string]interface{}{},
	}
}

func (r *Router) execMiddleware(ctx *Context) {
	for _, middleware := range r.Middleware {
		middleware.OnRequest(ctx)
	}
}
