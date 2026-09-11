package main

import (
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.md")
	if err := writeFileAtomic(p, []byte("hello")); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Fatalf("content = %q, want hello", got)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("leftover temp files: %v", entries)
	}
}

func TestHandleSave(t *testing.T) {
	dir := t.TempDir()
	s := &server{dir: dir, edit: true}

	req := httptest.NewRequest(http.MethodPost, "/__save?p=sub/a.md", strings.NewReader("# hi"))
	rec := httptest.NewRecorder()
	s.handleSave(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	got, err := os.ReadFile(filepath.Join(dir, "sub", "a.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "# hi" {
		t.Fatalf("content = %q", got)
	}
}

func TestFriendlyWriteErr(t *testing.T) {
	if got := friendlyWriteErr(nil); got != "" {
		t.Fatalf("nil error = %q, want empty", got)
	}
	if got := friendlyWriteErr(fs.ErrPermission); !strings.Contains(got, "read-only") {
		t.Fatalf("permission error = %q", got)
	}
	if got := friendlyWriteErr(errors.New("read-only file system")); !strings.Contains(got, "read-only") {
		t.Fatalf("read-only error = %q", got)
	}
	if got := friendlyWriteErr(errors.New("boom")); got != "boom" {
		t.Fatalf("other error = %q, want boom", got)
	}
}

func TestHandleSaveDisabled(t *testing.T) {
	s := &server{dir: t.TempDir(), edit: false}
	req := httptest.NewRequest(http.MethodPost, "/__save?p=a.md", strings.NewReader("x"))
	rec := httptest.NewRecorder()
	s.handleSave(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestHandleSaveClampsTraversal(t *testing.T) {
	dir := t.TempDir()
	s := &server{dir: dir, edit: true}
	req := httptest.NewRequest(http.MethodPost, "/__save?p=../evil.md", strings.NewReader("x"))
	rec := httptest.NewRecorder()
	s.handleSave(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(dir), "evil.md")); err == nil {
		t.Fatal("file escaped the serve directory")
	}
	if _, err := os.Stat(filepath.Join(dir, "evil.md")); err != nil {
		t.Fatalf("clamped file missing: %v", err)
	}
}

func TestHandleRender(t *testing.T) {
	s := &server{dir: t.TempDir(), edit: true}
	req := httptest.NewRequest(http.MethodPost, "/__render", strings.NewReader("# Title"))
	rec := httptest.NewRecorder()
	s.handleRender(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "<h1") {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestHandleFrag(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, dir, "a.md", "# A")
	s := &server{dir: dir, edit: true}

	req := httptest.NewRequest(http.MethodGet, "/__frag?p=a.md", nil)
	rec := httptest.NewRecorder()
	s.handleFrag(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "<html") {
		t.Fatal("fragment should not contain document shell")
	}
}
