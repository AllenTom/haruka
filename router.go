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
}

func NewRouter() *Router {
	return &Router{
		HandlerRouter: mux.NewRouter(),
		Handlers:      []RouterMapping{},
	}
}

func (r *Router) AddHandler(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		handler(&Context{
			Writer:     writer,
			Request:    request,
			Parameters: mux.Vars(request),
		})
	})
}

func (r *Router) GET(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		handler(&Context{
			Writer:     writer,
			Request:    request,
			Parameters: mux.Vars(request),
		})
	}).Methods("GET")
}

func (r *Router) POST(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		handler(&Context{
			Writer:     writer,
			Request:    request,
			Parameters: mux.Vars(request),
		})
	}).Methods("POST")
}

func (r *Router) PUT(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		handler(&Context{
			Writer:     writer,
			Request:    request,
			Parameters: mux.Vars(request),
		})
	}).Methods("PUT")
}

func (r *Router) PATCH(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		handler(&Context{
			Writer:     writer,
			Request:    request,
			Parameters: mux.Vars(request),
		})
	}).Methods("PATCH")
}

func (r *Router) DELETE(pattern string, handler RequestHandler) {
	r.HandlerRouter.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		handler(&Context{
			Writer:     writer,
			Request:    request,
			Parameters: mux.Vars(request),
		})
	}).Methods("DELETE")
}

func (r *Router) Static(pattern string, staticPath string) {
	r.HandlerRouter.PathPrefix(pattern).Handler(http.StripPrefix(pattern, http.FileServer(http.Dir(staticPath))))
}
