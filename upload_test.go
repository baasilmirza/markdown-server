package main

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleNew(t *testing.T) {
	dir := t.TempDir()
	s := &server{dir: dir, edit: true}

	form := url.Values{"path": {"sub/new"}, "content": {"# New"}}
	req := httptest.NewRequest(http.MethodPost, "/__new", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	s.handleNew(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	got, err := os.ReadFile(filepath.Join(dir, "sub", "new.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "# New" {
		t.Fatalf("content = %q", got)
	}

	form2 := url.Values{"path": {"sub/new"}, "content": {"again"}}
	req2 := httptest.NewRequest(http.MethodPost, "/__new", strings.NewReader(form2.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec2 := httptest.NewRecorder()
	s.handleNew(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("duplicate status = %d, want 200 form", rec2.Code)
	}
}

func TestHandleUpload(t *testing.T) {
	dir := t.TempDir()
	s := &server{dir: dir, edit: true}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("dir", "notes")
	part, err := mw.CreateFormFile("files", "up.md")
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte("# Up"))
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/__upload", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	s.handleUpload(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	got, err := os.ReadFile(filepath.Join(dir, "notes", "up.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "# Up" {
		t.Fatalf("content = %q", got)
	}
}

func TestHandleUploadDisabled(t *testing.T) {
	s := &server{dir: t.TempDir(), edit: false}
	req := httptest.NewRequest(http.MethodPost, "/__upload", nil)
	rec := httptest.NewRecorder()
	s.handleUpload(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestEditEnabled(t *testing.T) {
	cases := []struct {
		addr string
		mode string
		want bool
	}{
		{":8080", "auto", true},
		{"127.0.0.1:8080", "auto", true},
		{"localhost:8080", "auto", true},
		{"0.0.0.0:8080", "auto", false},
		{"192.168.1.5:8080", "auto", false},
		{"0.0.0.0:8080", "true", true},
		{"127.0.0.1:8080", "false", false},
	}
	for _, c := range cases {
		got, _ := editEnabled(c.addr, c.mode)
		if got != c.want {
			t.Errorf("editEnabled(%q, %q) = %v, want %v", c.addr, c.mode, got, c.want)
		}
	}
}
