package Haruka

import (
	"log"
	"net/http"
)

type Engine struct {
	Router *Router
}

func NewEngine() *Engine {
	return &Engine{
		Router: &Router{
			Handlers: []*RouterMapping{},
		},
	}
}
func (e *Engine) RunAndListen(addr string) {
	for _, handler := range e.Router.Handlers {
		http.HandleFunc(handler.Pattern, func(writer http.ResponseWriter, request *http.Request) {
			handler.HandlerFunc(&Context{
				Writer:  writer,
				Request: request,
			})
		})
	}
	log.Fatal(http.ListenAndServe(addr, nil))
}
