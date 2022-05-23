package trie

import (
	"net/http"
	"reflect"
	"testing"
)

type fakeHandler int

func (fakeHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

func TestTrie(t *testing.T) {
	trie := New()

	type search struct {
		method        string
		path          string
		expectedVars  Vars
		expectedError error
	}

	inserts := []struct {
		methods []string
		path    string
		tests   []search
	}{
		{
			path:    "/products",
			methods: []string{"GET", "POST"},
			tests: []search{
				{
					method:        "GET",
					path:          "/products",
					expectedVars:  Vars{},
					expectedError: nil,
				},
				{
					method:        "PATCH",
					path:          "/products",
					expectedVars:  Vars{},
					expectedError: ErrMethodNotFound,
				},
			},
		},
		{
			methods: []string{"GET", "PATCH"},
			path:    "/products/:pid",
			tests: []search{
				{
					method:        "GET",
					path:          "/products/1",
					expectedVars:  Vars{"pid": "1"},
					expectedError: nil,
				},
				{
					method:        "PATCH",
					path:          "/products/abc",
					expectedVars:  Vars{"pid": "abc"},
					expectedError: nil,
				},
				{
					method:        "PUT",
					path:          "/products/abc",
					expectedVars:  Vars{"pid": "abc"},
					expectedError: ErrMethodNotFound,
				},
			},
		},
		{
			methods: []string{"GET"},
			path:    "/products/carts",
			tests: []search{
				{
					method:        "GET",
					path:          "/products/carts",
					expectedVars:  Vars{},
					expectedError: nil,
				},
				{
					method:        "DELETE",
					path:          "/products/carts",
					expectedVars:  Vars{},
					expectedError: ErrMethodNotFound,
				},
			},
		},
		{
			methods: []string{"GET"},
			path:    "/products/:pid/stars",
			tests: []search{
				{
					method:        "GET",
					path:          "/products/100/stars",
					expectedVars:  Vars{"pid": "100"},
					expectedError: nil,
				},
				{
					method:        "GET",
					path:          "/products/xyz/stars",
					expectedVars:  Vars{"pid": "xyz"},
					expectedError: nil,
				},
			},
		},
		{
			path:    "/products/:pid/comments",
			methods: []string{"GET"},
			tests: []search{
				{
					method:        "GET",
					path:          "/products/200/comments",
					expectedVars:  Vars{"pid": "200"},
					expectedError: nil,
				},
				{
					method:        "GET",
					path:          "/products/xyz/comments",
					expectedVars:  Vars{"pid": "xyz"},
					expectedError: nil,
				},
			},
		},
		{
			path:    "/products/carts/:cid",
			methods: []string{"DELETE"},
			tests: []search{
				{
					method:        "DELETE",
					path:          "/products/carts/200",
					expectedVars:  Vars{"cid": "200"},
					expectedError: nil,
				},
				{
					method:        "DELETE",
					path:          "/products/carts/xyz",
					expectedVars:  Vars{"cid": "xyz"},
					expectedError: nil,
				},
			},
		},
		{
			path:    "/profiles/:id",
			methods: []string{"GET"},
			tests: []search{
				{
					method:        "GET",
					path:          "/profiles/100",
					expectedVars:  Vars{"id": "100"},
					expectedError: nil,
				},
				{
					method:        "POST",
					path:          "/profiles/xyz",
					expectedVars:  Vars{"id": "xyz"},
					expectedError: ErrMethodNotFound,
				},
			},
		},
		{
			path:    "/profiles/settings",
			methods: []string{"GET"},
			tests: []search{
				{
					method:        "GET",
					path:          "/profiles/settings",
					expectedVars:  Vars{},
					expectedError: nil,
				},
				{
					method:        "POST",
					path:          "/profiles/settings",
					expectedVars:  Vars{},
					expectedError: ErrMethodNotFound,
				},
			},
		},
	}

	// insert
	for i, r := range inserts {
		for _, m := range r.methods {
			assertNil(t, trie.InsertHandler(m, r.path, fakeHandler(i)))
		}
	}

	// insert with conflict path or method
	err := trie.InsertHandler("GET", "/profiles/settings", fakeHandler(99))
	assertNotNil(t, err)

	err = trie.InsertHandler("PUT", "/products/carts/:different_name/adc", fakeHandler(99))
	assertNotNil(t, err)

	// find
	for i, r := range inserts {
		exp := fakeHandler(i)
		for _, tc := range r.tests {
			got, vars, err := trie.FindHandler(tc.method, tc.path)
			if err != tc.expectedError {
				t.Fatalf("%s %s: expecting error %v; got %v", tc.method, tc.path, tc.expectedError, err)
			}

			if !reflect.DeepEqual(vars, tc.expectedVars) {
				t.Errorf("%s %s: expecting vars %v; got %v", tc.method, tc.path, tc.expectedVars, vars)
			}

			if err == nil && got != exp {
				t.Errorf("%s %s: expecting handler %v; got %v", tc.method, tc.path, exp, got)
			}
		}

	}

	// find unregistered paths
	_, _, err = trie.FindHandler("GET", "/profiles/abc/def")
	if err != ErrPathNotFound {
		t.Fatalf("expecting error %v; got %v", ErrMethodNotFound, err)
	}

	_, _, err = trie.FindHandler("GET", "/")
	if err != ErrPathNotFound {
		t.Fatalf("expecting error %v; got %v", ErrMethodNotFound, err)
	}

}

func assertNil(t *testing.T, err error) {
	if err != nil {
		t.Helper()
		t.Fatal(err)
	}
}

func assertNotNil(t *testing.T, err error) {
	if err == nil {
		t.Helper()
		t.Fatalf("expecting error not nil")
	}
}
