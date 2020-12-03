package haruka

import (
	"github.com/gorilla/mux"
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
	for _, handler := range e.Router.Handlers {
		e.Router.HandlerRouter.HandleFunc(handler.Pattern, func(writer http.ResponseWriter, request *http.Request) {
			handler.HandlerFunc(&Context{
				Writer:     writer,
				Request:    request,
				Parameters: mux.Vars(request),
			})
		})
	}
	mx := http.NewServeMux()
	mx.Handle("/", e.Router.HandlerRouter)
	server := &http.Server{
		Addr:    addr,
		Handler: mx,
	}
	e.server = server
	log.Fatal(server.ListenAndServe())
}

func (e *Engine) Close() {
	e.server.Close()
}
