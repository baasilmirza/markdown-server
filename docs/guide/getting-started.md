# Getting Started

Welcome. This page covers the shortest path from zero to a rendered page.

## Run the server

```powershell
go run . -dir .\docs -addr :8080
```

Then open <http://localhost:8080>.

## What happens on each request

1. The request path is cleaned and mapped to a file under `-dir`.
2. If the path has no extension, `.md` is appended.
3. The file is rendered with goldmark and wrapped in the page shell.

## Search me

The word **kangaroo** appears here so full-text search has a distinctive hit.
So does **kangaroo** again, on purpose.

## Notes

- Nested directories are supported, see [Advanced Usage](advanced.md).
- Back to the [index](../README.md).
