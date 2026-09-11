package main

import (
	"strings"
	"testing"
)

func TestSearchCatalog(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir, "one.md", "the kangaroo jumped")
	mustWrite(t, dir, "two.md", "nothing here")
	mustWrite(t, dir, "three.md", "kangaroo kangaroo kangaroo")

	cat, err := buildCatalog(dir)
	if err != nil {
		t.Fatal(err)
	}

	results := searchCatalog(cat, "kangaroo", 10)
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	if results[0].Rel != "three.md" {
		t.Fatalf("top hit = %q, want three.md", results[0].Rel)
	}
	for _, r := range results {
		if !strings.Contains(strings.ToLower(r.Snippet), "kangaroo") {
			t.Errorf("snippet for %s missing term: %q", r.Rel, r.Snippet)
		}
	}
}

func TestSearchCatalogEmptyAndLimit(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir, "a.md", "match")
	mustWrite(t, dir, "b.md", "match")

	cat, err := buildCatalog(dir)
	if err != nil {
		t.Fatal(err)
	}

	if got := searchCatalog(cat, "   ", 10); got != nil {
		t.Fatalf("blank query = %v, want nil", got)
	}
	if got := searchCatalog(cat, "match", 1); len(got) != 1 {
		t.Fatalf("limit 1 returned %d", len(got))
	}
}

func TestSnippet(t *testing.T) {
	if got := snippet("short text", 0); got != "short text" {
		t.Fatalf("snippet = %q", got)
	}
	long := strings.Repeat("x", 500) + "needle" + strings.Repeat("y", 500)
	got := snippet(long, strings.Index(long, "needle"))
	if !strings.Contains(got, "needle") {
		t.Fatalf("snippet missing needle: %q", got)
	}
}
