package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type server struct {
	dir string
}

func (s *server) handle(w http.ResponseWriter, r *http.Request) {
	rel := strings.TrimPrefix(path.Clean("/"+r.URL.Path), "/")
	if rel == "" || rel == "." {
		s.index(w, r)
		return
	}

	full := filepath.Join(s.dir, filepath.FromSlash(rel))
	if !within(s.dir, full) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
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
	body, err := renderMarkdown(src)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	title := strings.TrimSuffix(filepath.Base(full), ".md")
	s.render(w, title, route, body)
}

// resolve maps a request path to an absolute markdown file and its route.
func (s *server) resolve(rel string) (full, route string, ok bool) {
	if rel == "" || rel == "." {
		return "", "", false
	}
	full = filepath.Join(s.dir, filepath.FromSlash(rel))
	if !within(s.dir, full) {
		return "", "", false
	}
	if filepath.Ext(full) == "" {
		if _, err := os.Stat(full + ".md"); err == nil {
			full += ".md"
		}
	}
	if !strings.EqualFold(filepath.Ext(full), ".md") {
		return "", "", false
	}
	route = strings.TrimSuffix(filepath.ToSlash(rel), filepath.Ext(rel))
	return full, route, true
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	cat, err := buildCatalog(s.dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var b strings.Builder
	b.WriteString("<h1>Markdown files</h1>\n")
	if len(cat.docs) == 0 {
		b.WriteString("<p>No .md files found.</p>")
	} else {
		b.WriteString("<ul>\n")
		for _, d := range cat.docs {
			b.WriteString(`<li><a href="` + d.URL + `">` + template.HTMLEscapeString(d.Rel) + "</a></li>\n")
		}
		b.WriteString("</ul>\n")
	}

	renderPage(w, pageData{
		Title:     "Markdown files",
		SiteTitle: "Markdown Server",
		Items:     cat.sidebar(""),
		Body:      template.HTML(b.String()),
		Scripts:   pageScripts(),
	})
}

// handleSearch serves JSON full-text search results.
func (s *server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	cat, err := buildCatalog(s.dir)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	results := searchCatalog(cat, q, 20)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]any{
		"query":   q,
		"results": results,
	})
}

// render writes a document page with the recursive sidebar and prev/next links.
func (s *server) render(w http.ResponseWriter, title, route string, body []byte) {
	var items []sidebarItem
	var prev, next *doc
	if cat, err := buildCatalog(s.dir); err == nil {
		items = cat.sidebar(route)
		prev, next = cat.neighbors(route)
	}
	renderPage(w, pageData{
		Title:     title,
		SiteTitle: "Markdown Server",
		Items:     items,
		Body:      template.HTML(body),
		Prev:      prev,
		Next:      next,
		Scripts:   pageScripts(),
	})
}

// handleFrag returns just the rendered content for a document.
func (s *server) handleFrag(w http.ResponseWriter, r *http.Request) {
	rel := strings.TrimPrefix(path.Clean("/"+r.URL.Query().Get("p")), "/")
	full, _, ok := s.resolve(rel)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	src, err := os.ReadFile(full)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	body, err := renderMarkdown(src)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(body)
}

// handleTree returns the sidebar tree fragment for a route.
func (s *server) handleTree(w http.ResponseWriter, r *http.Request) {
	route := r.URL.Query().Get("route")
	cat, err := buildCatalog(s.dir)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := pageTmpl.ExecuteTemplate(w, "tree", pageData{Items: cat.sidebar(route)}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// writeJSONError writes a JSON error payload.
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// within reports whether target resolves inside root, following symlinks.
func within(root, target string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	if resolved, err := filepath.EvalSymlinks(rootAbs); err == nil {
		rootAbs = resolved
	}

	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	if resolved, err := filepath.EvalSymlinks(targetAbs); err == nil {
		targetAbs = resolved
	} else if resolved, err := filepath.EvalSymlinks(filepath.Dir(targetAbs)); err == nil {
		targetAbs = filepath.Join(resolved, filepath.Base(targetAbs))
	}

	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
