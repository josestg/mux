package mux_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/josestg/mux"
	"github.com/josestg/mux/internal/trie"
)

type fakeHandler int

type response struct {
	ID     int       `json:"id"`
	Method string    `json:"method"`
	Path   string    `json:"path"`
	Vars   trie.Vars `json:"vars"`
}

func (f fakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(&response{
		ID:     int(f),
		Method: r.Method,
		Path:   r.URL.Path,
		Vars:   mux.GetVars(r.Context()),
	})
}

func TestMux_ServeHTTP(t *testing.T) {

	handler := mux.New(mux.Default())

	handler.Handle(http.MethodGet, "/", fakeHandler(0))
	handler.HandleFunc(http.MethodPost, "/", func(w http.ResponseWriter, r *http.Request) { fakeHandler(1).ServeHTTP(w, r) })

	handler.Handle(http.MethodGet, "/a/b", fakeHandler(2))
	handler.Handle(http.MethodGet, "/a/:id", fakeHandler(3))

	handler.Handle(http.MethodGet, "/a/b/c", fakeHandler(4))
	handler.Handle(http.MethodGet, "/a/:id/c", fakeHandler(5))

	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		server.Close()
	})

	tests := []struct {
		method        string
		path          string
		exp           response
		wantErr       bool
		expStatusCode int
	}{
		{
			method: http.MethodGet,
			path:   "/",
			exp: response{
				ID:     0,
				Method: "GET",
				Path:   "/",
				Vars:   trie.Vars{},
			},
		},
		{
			method: http.MethodPost,
			path:   "/",
			exp: response{
				ID:     1,
				Method: "POST",
				Path:   "/",
				Vars:   trie.Vars{},
			},
		},

		{
			method: http.MethodGet,
			path:   "/a/b",
			exp: response{
				ID:     2,
				Method: "GET",
				Path:   "/a/b",
				Vars:   trie.Vars{},
			},
		},
		{
			method: http.MethodGet,
			path:   "/a/123",
			exp: response{
				ID:     3,
				Method: "GET",
				Path:   "/a/123",
				Vars:   trie.Vars{"id": "123"},
			},
		},

		{
			method: http.MethodGet,
			path:   "/a/b/c",
			exp: response{
				ID:     4,
				Method: "GET",
				Path:   "/a/b/c",
				Vars:   trie.Vars{},
			},
		},
		{
			method: http.MethodGet,
			path:   "/a/123/c",
			exp: response{
				ID:     5,
				Method: "GET",
				Path:   "/a/123/c",
				Vars:   trie.Vars{"id": "123"},
			},
		},
		{
			method:        "GET",
			path:          "/a/b/c/d",
			exp:           response{},
			wantErr:       true,
			expStatusCode: 404,
		},
		{
			method:        "PUT",
			path:          "/a/b/c",
			exp:           response{},
			wantErr:       true,
			expStatusCode: 405,
		},
	}

	client := server.Client()
	for _, tc := range tests {
		req, err := http.NewRequest(tc.method, server.URL+tc.path, nil)
		if err != nil {
			t.Fatal(err)
		}

		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if tc.wantErr {
			if res.StatusCode != tc.expStatusCode {
				t.Fatalf("expected %d; got %d", tc.expStatusCode, res.StatusCode)
			}
		} else {
			var got response
			_ = json.NewDecoder(res.Body).Decode(&got)
			if !reflect.DeepEqual(got, tc.exp) {
				t.Fatalf("expected %+v; got %+v", tc.exp, got)
			}
		}

	}

}
