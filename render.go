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
	Rel       string
	Dir       string
	Edit      bool
	Items     []sidebarItem
	Body      template.HTML
	Prev      *doc
	Next      *doc
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

// pageScripts returns the scripts injected into every page.
func pageScripts() template.HTML {
	return template.HTML(reloadScript + searchScript + navScript + themeScript)
}

const reloadScript = `<script>
(function () {
  var proto = location.protocol === 'https:' ? 'wss://' : 'ws://';

  function currentRoute() {
    return decodeURIComponent(location.pathname).replace(/^\/+/, '').replace(/\.md$/i, '');
  }

  function swap(html) {
    var el = document.getElementById('content');
    if (!el) { location.reload(); return; }
    var y = window.scrollY;
    el.innerHTML = html;
    window.scrollTo(0, y);
  }

  function refreshTree() {
    fetch('/__tree?route=' + encodeURIComponent(currentRoute()))
      .then(function (r) { return r.text(); })
      .then(function (html) {
        var nav = document.getElementById('tree');
        if (nav) { nav.innerHTML = html; }
      })
      .catch(function () {});
  }

  function connect() {
    var ws = new WebSocket(proto + location.host + '/ws');
    ws.onmessage = function (e) {
      var msg;
      try { msg = JSON.parse(e.data); } catch (err) { location.reload(); return; }
      if (!msg || msg.type !== 'reload') { return; }
      if (msg.tree) { refreshTree(); }
      var changed = (msg.path || '').replace(/\.md$/i, '');
      if (changed && changed === currentRoute()) {
        fetch('/__frag?p=' + encodeURIComponent(msg.path))
          .then(function (r) { if (!r.ok) { throw new Error('frag'); } return r.text(); })
          .then(swap)
          .catch(function () { location.reload(); });
      }
    };
    ws.onclose = function () { setTimeout(connect, 1000); };
  }
  connect();
})();
</script>`

const searchScript = `<script>
(function () {
  var input = document.getElementById('search');
  var box = document.getElementById('search-results');
  if (!input || !box) return;
  var timer = null;
  function clear() { box.innerHTML = ''; box.hidden = true; }
  function render(results) {
    box.innerHTML = '';
    if (!results.length) { box.hidden = true; return; }
    results.forEach(function (r) {
      var a = document.createElement('a');
      a.href = r.url;
      a.className = 'search-hit';
      var t = document.createElement('div');
      t.className = 'search-title';
      t.textContent = r.title;
      var s = document.createElement('div');
      s.className = 'search-snippet';
      s.textContent = r.snippet;
      a.appendChild(t);
      a.appendChild(s);
      box.appendChild(a);
    });
    box.hidden = false;
  }
  function run() {
    var q = input.value.trim();
    if (!q) { clear(); return; }
    fetch('/__search?q=' + encodeURIComponent(q))
      .then(function (r) { return r.json(); })
      .then(function (d) { render(d.results || []); })
      .catch(clear);
  }
  input.addEventListener('input', function () {
    clearTimeout(timer);
    timer = setTimeout(run, 150);
  });
  input.addEventListener('keydown', function (e) {
    if (e.key === 'Enter') {
      var first = box.querySelector('a');
      if (first) { location.href = first.getAttribute('href'); }
    }
    if (e.key === 'Escape') { input.value = ''; clear(); input.blur(); }
  });
  document.addEventListener('keydown', function (e) {
    if (e.key === '/' && document.activeElement !== input) {
      e.preventDefault();
      input.focus();
    }
  });
})();
</script>`

const navScript = `<script>
(function () {
  document.addEventListener('keydown', function (e) {
    var tag = e.target && e.target.tagName;
    if (tag === 'INPUT' || tag === 'TEXTAREA') return;
    var link = null;
    if (e.key === 'ArrowLeft' || e.key === 'k') {
      link = document.querySelector('a.nav-prev');
    } else if (e.key === 'ArrowRight' || e.key === 'j') {
      link = document.querySelector('a.nav-next');
    }
    if (link) { location.href = link.getAttribute('href'); }
  });
})();
</script>`

const themeScript = `<script>
(function () {
  var btn = document.getElementById('theme-toggle');
  if (!btn) return;
  function current() {
    return document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'light';
  }
  function label() {
    btn.textContent = current() === 'dark' ? 'Light mode' : 'Dark mode';
  }
  btn.addEventListener('click', function () {
    var next = current() === 'dark' ? 'light' : 'dark';
    document.documentElement.setAttribute('data-theme', next);
    try { localStorage.setItem('theme', next); } catch (e) {}
    label();
  });
  label();
})();
</script>`
