package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func mustWrite(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestBuildCatalogOrder(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir, "b.md", "b")
	mustWrite(t, dir, "a.md", "a")
	mustWrite(t, dir, "guide/z.md", "z")
	mustWrite(t, dir, "guide/a.md", "a")
	mustWrite(t, dir, "notes/2026/day.md", "day")
	mustWrite(t, dir, "ignore.txt", "nope")

	cat, err := buildCatalog(dir)
	if err != nil {
		t.Fatal(err)
	}

	var got []string
	for _, d := range cat.docs {
		got = append(got, d.Rel)
	}
	want := []string{"guide/a.md", "guide/z.md", "notes/2026/day.md", "a.md", "b.md"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("docs = %v, want %v", got, want)
	}
}

func TestCatalogNeighbors(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir, "guide/a.md", "a")
	mustWrite(t, dir, "guide/z.md", "z")
	mustWrite(t, dir, "notes/day.md", "d")

	cat, err := buildCatalog(dir)
	if err != nil {
		t.Fatal(err)
	}

	prev, next := cat.neighbors("guide/z")
	if prev == nil || prev.Rel != "guide/a.md" {
		t.Fatalf("prev = %v, want guide/a.md", prev)
	}
	if next == nil || next.Rel != "notes/day.md" {
		t.Fatalf("next = %v, want notes/day.md", next)
	}

	prev, next = cat.neighbors("guide/a")
	if prev != nil {
		t.Fatalf("first doc prev = %v, want nil", prev)
	}
	if next == nil || next.Rel != "guide/z.md" {
		t.Fatalf("first doc next = %v, want guide/z.md", next)
	}
}

func TestCatalogSidebarActive(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir, "guide/a.md", "a")
	mustWrite(t, dir, "b.md", "b")

	cat, err := buildCatalog(dir)
	if err != nil {
		t.Fatal(err)
	}

	items := cat.sidebar("guide/a")
	var dirs, active int
	for _, it := range items {
		if it.IsDir {
			dirs++
		}
		if it.Active {
			active++
			if it.Route != "guide/a" {
				t.Fatalf("active route = %q, want guide/a", it.Route)
			}
		}
	}
	if dirs != 1 {
		t.Fatalf("dirs = %d, want 1", dirs)
	}
	if active != 1 {
		t.Fatalf("active = %d, want 1", active)
	}
}

func TestRouteURL(t *testing.T) {
	cases := map[string]string{
		"":           "/",
		"a":          "/a",
		"guide/a":    "/guide/a",
		"a b/c":      "/a%20b/c",
		"notes/x&y":  "/notes/x&y",
		"q/why?what": "/q/why%3Fwhat",
		"n/a#frag":   "/n/a%23frag",
	}
	for in, want := range cases {
		if got := routeURL(in); got != want {
			t.Errorf("routeURL(%q) = %q, want %q", in, got, want)
		}
	}
}
