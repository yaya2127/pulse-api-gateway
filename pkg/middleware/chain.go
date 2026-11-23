package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

type Chain struct {
	middlewares []Middleware
}

func NewChain(middlewares ...Middleware) Chain {
	return Chain{middlewares: append(([]Middleware)(nil), middlewares...)}
}

func (c Chain) Then(h http.Handler) http.Handler {
	if h == nil {
		h = http.DefaultServeMux
	}
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}

func (c Chain) Append(m ...Middleware) Chain {
	newMiddlewares := make([]Middleware, 0, len(c.middlewares)+len(m))
	newMiddlewares = append(newMiddlewares, c.middlewares...)
	newMiddlewares = append(newMiddlewares, m...)
	return Chain{middlewares: newMiddlewares}
}
