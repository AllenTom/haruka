package haruka

import (
	"log"
	"net/http"
)

type Engine struct {
	Router *Router
	server *http.Server
}

func NewEngine() *Engine {
	return &Engine{
		Router: NewRouter(),
	}
}
func (e *Engine) RunAndListen(addr string) {
	log.Fatal(http.ListenAndServe(addr, e.Router.HandlerRouter))
}
