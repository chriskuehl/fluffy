package main

import (
	"bytes"
	"regexp"
	"testing"
)

func TestRegexHighlightFragment(t *testing.T) {
	tests := []struct {
		name    string
		regexp  string
		content string
		want    string
	}{
		{
			name:   "single match",
			regexp: "foo",
			content: `foo
bar
baz`,
			want: "L1",
		},
		{
			name:   "multiple matches",
			regexp: "foo",
			content: `foo
bar
foo
foo
baz
foo`,
			want: "L1,L3-4,L6",
		},
		{
			name:   "everything matches",
			regexp: ".",
			content: `foo
bar
foo
foo
baz
foo`,
			want: "L1-6",
		},
		{
			name:   "no matches",
			regexp: "qux",
			content: `foo
bar
baz`,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := regexHighlightFragment(regexp.MustCompile(tt.regexp), bytes.NewBufferString(tt.content))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
