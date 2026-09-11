package main

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// newPageData is the input to the new-file template.
type newPageData struct {
	Error string
}

// handleNew serves the new-file form and creates the file on submit.
func (s *server) handleNew(w http.ResponseWriter, r *http.Request) {
	if !s.requireEdit(w) {
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.renderNew(w, "")
	case http.MethodPost:
		s.createNew(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *server) createNew(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rel := strings.TrimSpace(r.FormValue("path"))
	content := r.FormValue("content")
	if len(content) > maxBody {
		s.renderNew(w, "Content is too large.")
		return
	}
	full, route, ok := s.resolveWrite(rel)
	if !ok {
		s.renderNew(w, "Enter a valid .md path inside the served directory.")
		return
	}
	if _, err := os.Stat(full); err == nil {
		s.renderNew(w, "A file already exists at that path.")
		return
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		s.renderNew(w, err.Error())
		return
	}
	if err := writeFileAtomic(full, []byte(content)); err != nil {
		s.renderNew(w, err.Error())
		return
	}
	http.Redirect(w, r, routeURL(route), http.StatusSeeOther)
}

func (s *server) renderNew(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := pageTmpl.ExecuteTemplate(w, "new", newPageData{Error: msg}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleUpload accepts one or more markdown files and writes them.
func (s *server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if !s.requireEdit(w) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	if err := r.ParseMultipartForm(maxBody); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	dir := strings.Trim(strings.TrimSpace(r.FormValue("dir")), "/")

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		writeJSONError(w, http.StatusBadRequest, "no files")
		return
	}

	var saved []string
	for _, fh := range files {
		name := filepath.Base(fh.Filename)
		if name == "" || name == "." || strings.HasPrefix(name, ".") {
			continue
		}
		if ext := filepath.Ext(name); ext != "" && !strings.EqualFold(ext, ".md") {
			continue
		}
		rel := name
		if dir != "" {
			rel = dir + "/" + name
		}
		full, route, ok := s.resolveWrite(rel)
		if !ok {
			continue
		}
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			continue
		}
		src, err := fh.Open()
		if err != nil {
			continue
		}
		data, err := io.ReadAll(io.LimitReader(src, maxBody))
		src.Close()
		if err != nil {
			continue
		}
		if err := writeFileAtomic(full, data); err != nil {
			continue
		}
		saved = append(saved, route+".md")
	}

	if strings.Contains(r.Header.Get("Accept"), "application/json") {
		writeJSON(w, map[string]any{"ok": true, "saved": saved})
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
