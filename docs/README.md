# Sample Documentation

This directory holds sample markdown used to exercise every feature of
markdown-server. It is intentionally nested so the sidebar tree, search,
prev/next navigation, and the editor all have something to chew on.

## Sections

- [Getting Started](guide/getting-started.md) - install and first run.
- [Advanced Usage](guide/advanced.md) - watcher, flags, deployment.
- [API Reference](reference/api.md) - HTTP routes and message shapes.
- [FAQ](reference/faq.md) - common questions.
- [Daily Notes](notes/2026/day-1.md) - dated entries.
- [Long Document](long-doc.md) - a long page used to test scroll preservation.

## Quick taste

| Feature | Keyword to search |
| --- | --- |
| Sidebar tree | `recursive` |
| Full-text search | `kangaroo` |
| Prev / next | `navigation` |
| Theme toggle | `theme` |
| Incremental reload | `scroll` |
| Editor | `editor` |
| Upload | `upload` |

```go
package main

import "fmt"

func main() {
	fmt.Println("hello from the sample docs")
}
```
