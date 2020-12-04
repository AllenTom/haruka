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
