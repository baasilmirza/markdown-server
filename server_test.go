package main

import (
	"path/filepath"
	"testing"
)

func TestWithin(t *testing.T) {
	root := t.TempDir()

	if !within(root, filepath.Join(root, "a.md")) {
		t.Error("file inside root reported outside")
	}
	if within(root, filepath.Join(root, "..", "outside.md")) {
		t.Error("path outside root reported inside")
	}
	if !within(root, filepath.Join(root, "sub", "new.md")) {
		t.Error("non-existent file with in-root parent reported outside")
	}
}

func TestResolve(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir, "guide/a.md", "a")
	s := &server{dir: dir}

	full, route, ok := s.resolve("guide/a")
	if !ok || route != "guide/a" {
		t.Fatalf("resolve = %q, %q, %v", full, route, ok)
	}
	if _, _, ok := s.resolve("guide/missing"); ok {
		t.Error("missing file resolved")
	}
	if _, _, ok := s.resolve("nope.txt"); ok {
		t.Error("non-markdown resolved")
	}
}

func TestResolveWrite(t *testing.T) {
	dir := t.TempDir()
	s := &server{dir: dir}

	full, route, ok := s.resolveWrite("notes/new")
	if !ok || route != "notes/new" {
		t.Fatalf("resolveWrite = %q, %q, %v", full, route, ok)
	}
	if filepath.Ext(full) != ".md" {
		t.Fatalf("resolveWrite ext = %q, want .md", filepath.Ext(full))
	}
	full2, _, ok := s.resolveWrite("../escape")
	if !ok {
		t.Fatal("clamped path rejected")
	}
	if filepath.Dir(full2) != dir {
		t.Fatalf("clamped path dir = %q, want %q", filepath.Dir(full2), dir)
	}
	if _, _, ok := s.resolveWrite(""); ok {
		t.Error("empty path accepted")
	}
}

func TestDocDir(t *testing.T) {
	cases := map[string]string{
		"a":         "",
		"guide/a":   "guide",
		"notes/x/y": "notes/x",
	}
	for in, want := range cases {
		if got := docDir(in); got != want {
			t.Errorf("docDir(%q) = %q, want %q", in, got, want)
		}
	}
}
