package main

import (
	"math"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	cases := []struct {
		name    string
		a, b    string
		want    float64
		wantErr bool
	}{
		{"identical vectors", `[1,0,0]`, `[1,0,0]`, 1, false},
		{"opposite vectors", `[1,0,0]`, `[-1,0,0]`, -1, false},
		{"orthogonal vectors", `[1,0]`, `[0,1]`, 0, false},
		{"mismatched length", `[1,0]`, `[1,0,0]`, 0, true},
		{"empty vectors", `[]`, `[]`, 0, true},
		{"invalid json", `not json`, `[1,0]`, 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := cosineSimilarity(c.a, c.b)
			if c.wantErr {
				if err == nil {
					t.Fatalf("cosineSimilarity(%q, %q) = %v, want error", c.a, c.b, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("cosineSimilarity(%q, %q) unexpected error: %v", c.a, c.b, err)
			}
			if math.Abs(got-c.want) > 1e-9 {
				t.Errorf("cosineSimilarity(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
			}
		})
	}
}

func TestCosineSimilarityZeroNorm(t *testing.T) {
	got, err := cosineSimilarity(`[0,0,0]`, `[1,2,3]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Errorf("cosineSimilarity with zero-norm vector = %v, want 0", got)
	}
}
