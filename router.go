package Haruka

type RouterMapping struct {
	Pattern     string
	HandlerFunc RequestHandler
}
type Router struct {
	Handlers []*RouterMapping
}

func (r *Router) AddHandler(pattern string, handler RequestHandler) {
	r.Handlers = append(r.Handlers, &RouterMapping{
		Pattern:     pattern,
		HandlerFunc: handler,
	})
}
