package mux

import (
	"net/http"
)

type MiddlewareFunc func(handler http.Handler) http.Handler

type Middleware interface {
	Middleware(handler http.Handler) http.Handler
}

func (mw MiddlewareFunc) Middleware(handler http.Handler) http.Handler {
	return mw(handler)
}

func (m *Mux) Use(mws ...MiddlewareFunc) {
	for _, mw := range mws {
		m.middlewares = append(m.middlewares, mw)
	}
}
