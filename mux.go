package mux

import (
	"context"
	"net/http"

	"github.com/josestg/mux/internal/trie"
)

// Mux is an HTTP request multiplexer. It matches the URL of each incoming
// request against a list of registered patterns and calls the handler for the
// pattern that matches the URL.
type Mux struct {
	router      *trie.Trie
	options     *Options
	middlewares []Middleware
}

// New creates a new Mux with Default option.
func New(appliers ...OptionApplier) *Mux {
	options := newDefaultOption()

	for _, apply := range appliers {
		apply(options)
	}

	return &Mux{
		router:      trie.New(),
		options:     options,
		middlewares: []Middleware{},
	}
}

// Handle registers the http.Handler for the given HTTP method and URL path.
func (m *Mux) Handle(method string, path string, handler http.Handler) {
	if err := m.router.InsertHandler(method, path, handler); err != nil {
		panic(err)
	}
}

// HandleFunc registers the http.HandlerFunc for the given HTTP method
// and URL path.
func (m *Mux) HandleFunc(method string, path string, handlerFunc http.HandlerFunc) {
	m.Handle(method, path, handlerFunc)
}

// ServeHTTP implements the http.Handler interface.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, vars, err := m.router.FindHandler(r.Method, r.URL.Path)
	if err != nil {
		switch err {
		case trie.ErrMethodNotFound:
			handler = m.options.MethodNotFoundHandler
		case trie.ErrPathNotFound:
			handler = m.options.RoutesNotFoundHandler
		}
	}

	wrappedHandler := handler
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		wrappedHandler = m.middlewares[i].Middleware(wrappedHandler)
	}

	ctx := contextWithVars(r.Context(), vars)
	wrappedHandler.ServeHTTP(w, r.WithContext(ctx))
}

func (m *Mux) useMiddleware(mw Middleware) {
	m.middlewares = append(m.middlewares, mw)
}

type contextType struct{}

var (
	varsContextKey = new(contextType)
)

func contextWithVars(ctx context.Context, vars trie.Vars) context.Context {
	return context.WithValue(ctx, varsContextKey, vars)
}

// GetVars returns URL variables.
func GetVars(ctx context.Context) trie.Vars {
	vars, _ := ctx.Value(varsContextKey).(trie.Vars)
	return vars
}

// OptionApplier is a function for applying option.
type OptionApplier func(o *Options)

// Options holds Mux optional fields.
type Options struct {
	RoutesNotFoundHandler http.Handler
	MethodNotFoundHandler http.Handler
}

// Default is a default option applier.
func Default() OptionApplier {
	return func(o *Options) {
		o.RoutesNotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		})
		o.MethodNotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		})
	}
}

func newDefaultOption() *Options {
	var options Options
	Default()(&options)
	return &options
}
