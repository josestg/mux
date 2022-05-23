package trie

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var (
	ErrPathNotFound   = errors.New("path is not found")
	ErrMethodNotFound = errors.New("method not found")
)

// Trie is inspired by the Trie data structure (prefix tree) which is used
// to find or store a specific handler using a method and URL pair.
//
// See: https://en.wikipedia.org/wiki/Trie.
type Trie struct {
	root *node
}

// New creates a new Trie.
func New() *Trie {
	return &Trie{
		root: newNode(rootLabel),
	}
}

// InsertHandler inserts a new handler.
func (t *Trie) InsertHandler(method string, path string, handler http.Handler) error {
	tokens := t.tokenizePath(path)

	p := t.root
	for _, v := range tokens {
		switch v.kind {
		case route:
			if _, ok := p.children[v.value]; !ok {
				p.children[v.value] = newNode(v.value)
			}
			p = p.children[v.value]
		case variable:
			child, exists := p.children[varsLabel]
			switch {
			case !exists:
				p.children[varsLabel] = newNode(v.value)
				p = p.children[varsLabel]
			case exists && child.label != v.value:
				return fmt.Errorf("variable name is differs with the previously registered. want=(%s), got=(%s)", child.label, v.value)
			case exists && child.label == v.value:
				p = child
			}
		}
	}

	if ok := p.handlers.set(method, handler); !ok {
		return fmt.Errorf("conflict handler. %s %s already has a handler", method, path)
	}

	return nil
}

// FindHandler finds a handler.
func (t *Trie) FindHandler(method string, path string) (http.Handler, Vars, error) {
	vars := make(Vars)
	tokens := t.tokenizePath(path)

	p := t.root
	for _, v := range tokens {
		nextNode, exists := p.children[v.value]
		if !exists {
			child, exists := p.children[varsLabel]
			if !exists {
				return nil, vars, ErrPathNotFound
			}

			p = child
			vars[child.label] = strings.TrimPrefix(v.value, "/")
			continue
		}

		p = nextNode
	}

	handler, exists := p.handlers.get(method)
	if !exists {
		return nil, vars, ErrMethodNotFound
	}

	return handler, vars, nil
}

func (t *Trie) tokenizePath(p string) []token {
	return tokenizePath(segmentizePath(cleanPath(p)))
}

type Vars map[string]string

func (v Vars) Get(name string) string {
	val, _ := v[name]
	return val
}

type handlers map[string]http.Handler

func (h handlers) method(m string) string {
	return strings.ToUpper(m)
}

func (h handlers) set(method string, handler http.Handler) bool {
	method = h.method(method)
	if _, found := h[method]; found {
		return false
	}

	h[method] = handler
	return true
}

func (h handlers) get(method string) (http.Handler, bool) {
	method = h.method(method)
	handler, found := h[method]
	return handler, found
}

const (
	rootLabel = "<root>"
	varsLabel = "<vars>"
)

type node struct {
	label    string
	handlers handlers
	children map[string]*node
}

func newNode(label string) *node {
	return &node{
		label:    label,
		handlers: make(handlers),
		children: make(map[string]*node),
	}
}
