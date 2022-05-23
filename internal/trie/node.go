package trie

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
