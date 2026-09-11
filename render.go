package main

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var markdown = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
)

// renderMarkdown converts markdown source to an HTML fragment.
func renderMarkdown(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := markdown.Convert(src, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// pageData is the input to the page template.
type pageData struct {
	Title     string
	SiteTitle string
	Items     []sidebarItem
	Body      template.HTML
	Scripts   template.HTML
}

// renderPage writes the full HTML document for data.
func renderPage(w http.ResponseWriter, data pageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := pageTmpl.ExecuteTemplate(w, "page", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

const reloadScript = `<script>
(function () {
  var proto = location.protocol === 'https:' ? 'wss://' : 'ws://';
  function connect() {
    var ws = new WebSocket(proto + location.host + '/ws');
    ws.onmessage = function (e) {
      if (e.data === 'reload') { location.reload(); }
    };
    ws.onclose = function () { setTimeout(connect, 1000); };
  }
  connect();
})();
</script>`
