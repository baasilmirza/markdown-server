# API Reference

All routes are served by the same process.

## Read routes

| Route | Method | Description |
| --- | --- | --- |
| `/` | GET | Index / first document |
| `/<path>` | GET | Render `<path>.md` |
| `/__frag?p=<path>` | GET | Rendered content fragment only |
| `/__tree` | GET | Sidebar tree fragment |
| `/__search?q=<term>` | GET | JSON search results |
| `/ws` | GET | WebSocket for live reload |

## Write routes (require `-edit`)

| Route | Method | Description |
| --- | --- | --- |
| `/__edit?p=<path>` | GET | Editor for a document |
| `/__render` | POST | Render markdown to HTML preview |
| `/__save?p=<path>` | POST | Save raw markdown |
| `/__new` | GET/POST | Create a new document |
| `/__upload` | POST | Upload one or more `.md` files |

## Live reload message

```json
{ "type": "reload", "path": "guide/advanced.md", "op": "write" }
```

`op` is one of `write`, `create`, `remove`, or `rename`. A `create` or `remove`
also triggers a sidebar refresh.

## Errors

```json
{ "error": "forbidden" }
```
