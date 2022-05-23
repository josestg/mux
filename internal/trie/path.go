package trie

import (
	"path"
	"strings"
)

const (
	route = iota
	variable
)

type token struct {
	kind  int
	value string
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}

	if p[0] != '/' {
		p = "/" + p
	}

	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		// Fast path for common case of p being the string we want:
		if len(p) == len(np)+1 && strings.HasPrefix(p, np) {
			np = p
		} else {
			np += "/"
		}
	}

	return np
}

func segmentizePath(p string) []string {
	p = strings.TrimSuffix(p, "/")
	p = strings.TrimPrefix(p, "/")

	parts := strings.Split(p, "/")
	for i, v := range parts {
		parts[i] = "/" + v
	}

	return parts
}

func tokenizePath(segments []string) []token {
	tokens := make([]token, 0, len(segments))

	for _, v := range segments {
		if strings.HasPrefix(v, "/:") {
			tokens = append(tokens, token{
				kind:  variable,
				value: strings.TrimPrefix(v, "/:"),
			})
		} else {
			tokens = append(tokens, token{
				kind:  route,
				value: v,
			})
		}
	}

	return tokens
}
