package haruka

// Group is a collection of routes sharing a common prefix and a
// common middleware chain. Groups can be nested; the effective chain
// is the concatenation of all ancestor chains (outermost first).
type Group struct {
	router      *Router
	prefix      string
	middlewares []Middleware
}

// Group creates a sub group with an extended prefix and additional
// middlewares.
func (g *Group) Group(prefix string, middlewares ...Middleware) *Group {
	return &Group{
		router:      g.router,
		prefix:      g.prefix + prefix,
		middlewares: g.chain(middlewares),
	}
}

// UseMiddleware appends middlewares to the group chain. They only
// apply to routes registered after this call.
func (g *Group) UseMiddleware(middlewares ...Middleware) {
	g.middlewares = append(g.middlewares, middlewares...)
}

// chain merges current group chain with extra middlewares.
func (g *Group) chain(extra []Middleware) []Middleware {
	merged := make([]Middleware, 0, len(g.middlewares)+len(extra))
	merged = append(merged, g.middlewares...)
	merged = append(merged, extra...)
	return merged
}

// fullPattern joins group prefix with route pattern.
func (g *Group) fullPattern(pattern string) string {
	return g.prefix + pattern
}

func (g *Group) AddHandler(pattern string, handler RequestHandler) {
	g.router.register(g.fullPattern(pattern), nil, handler, g.middlewares)
}

func (g *Group) GET(pattern string, handler RequestHandler) {
	g.router.register(g.fullPattern(pattern), []string{"GET"}, handler, g.middlewares)
}

func (g *Group) POST(pattern string, handler RequestHandler) {
	g.router.register(g.fullPattern(pattern), []string{"POST"}, handler, g.middlewares)
}

func (g *Group) PUT(pattern string, handler RequestHandler) {
	g.router.register(g.fullPattern(pattern), []string{"PUT"}, handler, g.middlewares)
}

func (g *Group) PATCH(pattern string, handler RequestHandler) {
	g.router.register(g.fullPattern(pattern), []string{"PATCH"}, handler, g.middlewares)
}

func (g *Group) DELETE(pattern string, handler RequestHandler) {
	g.router.register(g.fullPattern(pattern), []string{"DELETE"}, handler, g.middlewares)
}

func (g *Group) METHODS(pattern string, methods []string, handler RequestHandler) {
	g.router.register(g.fullPattern(pattern), methods, handler, g.middlewares)
}
