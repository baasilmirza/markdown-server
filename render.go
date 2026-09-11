package main

import (
	"bytes"
	"fmt"

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

const pageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%s</title>
<style>
:root { color-scheme: light dark; }
body {
  max-width: 46rem;
  margin: 0 auto;
  padding: 2rem 1rem;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  line-height: 1.6;
}
pre { background: rgba(127,127,127,.15); padding: 1rem; overflow-x: auto; border-radius: 6px; }
code { background: rgba(127,127,127,.15); padding: .1em .3em; border-radius: 4px; }
pre code { background: none; padding: 0; }
table { border-collapse: collapse; }
th, td { border: 1px solid rgba(127,127,127,.4); padding: .4rem .7rem; }
blockquote { border-left: 3px solid rgba(127,127,127,.5); margin-left: 0; padding-left: 1rem; color: inherit; }
</style>
</head>
<body>
%s
%s
</body>
</html>
`

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

// page wraps a rendered markdown fragment in the HTML document shell.
func page(title string, body []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf(pageTemplate, title, body, reloadScript))
	return buf.Bytes()
}
