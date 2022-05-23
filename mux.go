package mux

import (
	"context"
	"net/http"

	"github.com/josestg/mux/internal/trie"
)

type Mux struct {
	router  *trie.Trie
	options *Options
}

func New(appliers ...OptionApplier) *Mux {
	options := newDefaultOption()

	for _, apply := range appliers {
		apply(options)
	}

	return &Mux{
		router:  trie.New(),
		options: options,
	}
}

func (m *Mux) Handle(method string, path string, handler http.Handler) {
	if err := m.router.InsertHandler(method, path, handler); err != nil {
		panic(err)
	}
}

func (m *Mux) HandleFunc(method string, path string, handlerFunc http.HandlerFunc) {
	m.Handle(method, path, handlerFunc)
}

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

	ctx := contextWithVars(r.Context(), vars)
	handler.ServeHTTP(w, r.WithContext(ctx))
}

type contextType struct{}

var (
	varsContextKey = new(contextType)
)

func contextWithVars(ctx context.Context, vars trie.Vars) context.Context {
	return context.WithValue(ctx, varsContextKey, vars)
}

func GetVars(ctx context.Context) trie.Vars {
	vars, _ := ctx.Value(varsContextKey).(trie.Vars)
	return vars
}

type OptionApplier func(o *Options)

type Options struct {
	RoutesNotFoundHandler http.Handler
	MethodNotFoundHandler http.Handler
}

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
