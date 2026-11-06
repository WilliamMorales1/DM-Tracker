package main

import "testing"

func TestFtsMatchQuery(t *testing.T) {
	cases := []struct {
		name  string
		query string
		want  string
	}{
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
		{"single term", "goblin", `"goblin"*`},
		{"multiple terms", "drow priestess", `"drow"* OR "priestess"*`},
		{"keeps apostrophes, strips other punctuation", "goblin's lair!", `"goblin's"* OR "lair"*`},
		{"escapes embedded quotes", `say "hello"`, `"say"* OR "hello"*`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ftsMatchQuery(c.query)
			if got != c.want {
				t.Errorf("ftsMatchQuery(%q) = %q, want %q", c.query, got, c.want)
			}
		})
	}
}
