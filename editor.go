package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// maxBody caps the size of markdown accepted by write routes.
const maxBody = 5 << 20

// editorData is the input to the editor template.
type editorData struct {
	Rel     string
	Route   string
	URL     string
	Title   string
	Content string
}

// requireEdit reports whether write routes are enabled, writing 403 if not.
func (s *server) requireEdit(w http.ResponseWriter) bool {
	if !s.edit {
		writeJSONError(w, http.StatusForbidden, "editing disabled")
		return false
	}
	return true
}

// handleEdit renders the in-browser editor for a document.
func (s *server) handleEdit(w http.ResponseWriter, r *http.Request) {
	if !s.requireEdit(w) {
		return
	}
	rel := strings.TrimPrefix(path.Clean("/"+r.URL.Query().Get("p")), "/")
	full, route, ok := s.resolve(rel)
	if !ok {
		http.NotFound(w, r)
		return
	}
	src, err := os.ReadFile(full)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	relRoot, err := filepath.Rel(s.dir, full)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	relRoot = filepath.ToSlash(relRoot)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := pageTmpl.ExecuteTemplate(w, "editor", editorData{
		Rel:     relRoot,
		Route:   route,
		URL:     routeURL(route),
		Title:   strings.TrimSuffix(filepath.Base(full), ".md"),
		Content: string(src),
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// handleRender renders raw markdown to an HTML preview.
func (s *server) handleRender(w http.ResponseWriter, r *http.Request) {
	if !s.requireEdit(w) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	src, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	body, err := renderMarkdown(src)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(body)
}

// handleSave writes raw markdown to a document, atomically.
func (s *server) handleSave(w http.ResponseWriter, r *http.Request) {
	if !s.requireEdit(w) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	full, route, ok := s.resolveWrite(r.URL.Query().Get("p"))
	if !ok {
		writeJSONError(w, http.StatusBadRequest, "invalid path")
		return
	}
	src, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		writeJSONError(w, http.StatusInternalServerError, friendlyWriteErr(err))
		return
	}
	if err := writeFileAtomic(full, src); err != nil {
		writeJSONError(w, http.StatusInternalServerError, friendlyWriteErr(err))
		return
	}
	writeJSON(w, map[string]any{"ok": true, "path": route + ".md"})
}

// friendlyWriteErr turns filesystem write errors into a readable message.
func friendlyWriteErr(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, fs.ErrPermission) || strings.Contains(strings.ToLower(err.Error()), "read-only") {
		return "serve directory is read-only; mount it read-write to enable editing"
	}
	return err.Error()
}

// resolveWrite maps a request path to a writable markdown file, appending .md
// when missing and rejecting anything outside the serve directory.
func (s *server) resolveWrite(rel string) (full, route string, ok bool) {
	rel = strings.TrimPrefix(path.Clean("/"+rel), "/")
	if rel == "" || rel == "." {
		return "", "", false
	}
	if !strings.EqualFold(filepath.Ext(rel), ".md") {
		rel += ".md"
	}
	full = filepath.Join(s.dir, filepath.FromSlash(rel))
	if !within(s.dir, full) {
		return "", "", false
	}
	route = strings.TrimSuffix(filepath.ToSlash(rel), filepath.Ext(rel))
	return full, route, true
}

// writeFileAtomic writes data to path via a temp file and rename so readers
// never observe a partial write.
func writeFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".md-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// writeJSON writes v as a JSON response.
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
}
