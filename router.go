package haruka

import "github.com/gorilla/mux"

type RouterMapping struct {
	Pattern     string
	HandlerFunc RequestHandler
}

type Router struct {
	Handlers      []*RouterMapping
	HandlerRouter *mux.Router
}

func NewRouter() *Router {
	return &Router{
		HandlerRouter: mux.NewRouter(),
		Handlers:      []*RouterMapping{},
	}
}

func (r *Router) AddHandler(pattern string, handler RequestHandler) {
	r.Handlers = append(r.Handlers, &RouterMapping{
		Pattern:     pattern,
		HandlerFunc: handler,
	})
}
