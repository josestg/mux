package mux

import (
	"encoding/json"
	"net/http"
	"testing"
)

type middlewareTest struct {
	executedMiddlewares int
}

func (m *middlewareTest) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.executedMiddlewares++
		handler.ServeHTTP(w, r)
	})
}

func TestMiddleware(t *testing.T) {
	m := New()
	mt := &middlewareTest{}

	t.Run("test add middleware", func(t *testing.T) {
		m.useMiddleware(mt)
		if len(m.middlewares) != 1 || m.middlewares[0] != mt {
			t.Errorf("expected %v middlewares, got %v middlewares", 1, len(m.middlewares))
		}

		m.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// do middleware stuff here
			})
		})

		if len(m.middlewares) != 2 {
			t.Errorf("expected %v middlewares, got %v middlewares", 2, len(m.middlewares))
		}

		m.Use(mt.Middleware)
		if len(m.middlewares) != 3 {
			t.Errorf("expected %v middlewares, got %v middlewares", 3, len(m.middlewares))
		}
	})

	t.Run("test router middleware", func(t *testing.T) {
		m.useMiddleware(mt)

		t.Run("test handler middleware", func(t *testing.T) {
			m.HandleFunc(http.MethodGet, "/ping", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]string{"message": http.StatusText(http.StatusOK)})

				m.ServeHTTP(w, r)

				if mt.executedMiddlewares != 4 {
					t.Fatalf("Expected %d calls, but got only %d", 2, len(m.middlewares))
				}
			})

			if len(m.middlewares) != 4 {
				t.Fatalf("Expected %d calls, but got only %d", 2, len(m.middlewares))
			}
		})
	})
}
