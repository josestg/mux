package trie

import (
	"reflect"
	"testing"
)

func Test_CleanPath(t *testing.T) {
	type args struct {
		p string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "given an empty path. expecting got path = /",
			args: args{
				p: "",
			},
			want: "/",
		},
		{
			name: "given path without prefix /. expecting got path with prefix /",
			args: args{
				p: "a/b/c",
			},
			want: "/a/b/c",
		},
		{
			name: "given path with prefix / and suffix /. expecting got a same path.",
			args: args{
				p: "/a/b/c/",
			},
			want: "/a/b/c/",
		},
		{
			name: "given path redundant slashes. expecting a clean path",
			args: args{
				p: "//a/b////{c}//",
			},
			want: "/a/b/{c}/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanPath(tt.args.p); got != tt.want {
				t.Errorf("cleanPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_SegmentizePath(t *testing.T) {
	tests := []struct {
		p    string
		want []string
	}{
		{
			p:    "/",
			want: []string{"/"},
		},
		{
			p:    ":a",
			want: []string{"/:a"},
		},
		{
			p:    "/a/b/c",
			want: []string{"/a", "/b", "/c"},
		},
		{
			p:    "/a/:b/c",
			want: []string{"/a", "/:b", "/c"},
		},
	}

	for _, tt := range tests {
		got := segmentizePath(tt.p)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("segmentizePath() = %v, want %v", got, tt.want)
		}
	}
}

func Test_TokenizePath(t *testing.T) {
	tests := []struct {
		p    []string
		want []token
	}{
		{
			p:    []string{"/a", "/b"},
			want: []token{{kind: route, value: "/a"}, {kind: route, value: "/b"}},
		},
		{
			p:    []string{"/a", "/:b"},
			want: []token{{kind: route, value: "/a"}, {kind: variable, value: "b"}},
		},
	}

	for _, tt := range tests {
		got := tokenizePath(tt.p)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("segmentizePath() = %v, want %v", got, tt.want)
		}
	}
}
