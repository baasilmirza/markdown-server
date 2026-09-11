package main

import (
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
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
	if filepath.Ext(full) == "" {
		if _, err := os.Stat(full + ".md"); err == nil {
			full += ".md"
		}
	}
	if !strings.EqualFold(filepath.Ext(full), ".md") {
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(page(title, body))
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.EqualFold(filepath.Ext(e.Name()), ".md") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString("<h1>Markdown files</h1>\n")
	if len(names) == 0 {
		b.WriteString("<p>No .md files found.</p>")
	} else {
		b.WriteString("<ul>\n")
		for _, n := range names {
			href := "/" + url.PathEscape(strings.TrimSuffix(n, filepath.Ext(n)))
			b.WriteString(`<li><a href="` + href + `">` + template.HTMLEscapeString(n) + "</a></li>\n")
		}
		b.WriteString("</ul>\n")
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(page("Markdown files", []byte(b.String())))
}

// within reports whether target resolves inside root.
func within(root, target string) bool {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return false
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
