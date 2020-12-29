package haruka

import (
	"fmt"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

type Engine struct {
	Router      *Router
	Middlewares []Middleware
	server      *http.Server
	Logger      *logrus.Logger
	Cros        *cors.Cors
}

func NewEngine() *Engine {
	return &Engine{
		Router:      NewRouter(),
		Middlewares: make([]Middleware, 0),
		Logger:      logrus.New(),
	}
}

func (e *Engine) UseMiddleware(middleware Middleware) {
	e.Middlewares = append(e.Middlewares, middleware)
}
func (e *Engine) UseCors(cors *cors.Cors) {
	e.Cros = cors
}
func (e *Engine) RunAndListen(addr string) {
	e.Router.Middleware = e.Middlewares
	e.Logger.Info(fmt.Sprintf("application run in %s", addr))
	var router http.Handler
	router = e.Router.HandlerRouter
	if e.Cros != nil {
		router = e.Cros.Handler(e.Router.HandlerRouter)
	}
	log.Fatal(http.ListenAndServe(addr, router))
}
