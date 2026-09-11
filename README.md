# markdown-server

Serves `.md` files rendered to HTML and live-reloads the browser over a
WebSocket whenever a file changes. Includes a recursive sidebar, full-text
search, prev/next navigation, a light/dark theme toggle, incremental reload,
and an in-browser editor with upload and new-file support.

## Features

- Recursive file tree sidebar with the active document highlighted.
- Server-side full-text search with snippets (press `/` to focus).
- Prev/next navigation in tree order (arrow keys, or `j`/`k`).
- Light/dark theme toggle, remembered per browser.
- Incremental reload: editing the current document swaps its content and
  preserves scroll; adding or removing a file refreshes the sidebar.
- In-browser editor with a server-rendered preview (`Ctrl+S` to save).
- Upload one or more `.md` files and create new documents.

## Run with Docker Compose

```powershell
docker compose up --build
# open http://localhost:8080
```

Edit any file in `./docs`; the page reloads automatically.

## Run with Docker (no compose)

```powershell
docker build -t markdown-server .
docker run --rm -p 8080:8080 -v "${PWD}\docs:/docs:rw" markdown-server
```

The mount is read-write so the editor, upload, and new-file routes work.
On Linux, add `--user "$(id -u):$(id -g)"` so the container can write the
bind-mounted files. With Compose, set `DOCKER_UID` and `DOCKER_GID`:

```sh
export DOCKER_UID=$(id -u) DOCKER_GID=$(id -g)
docker compose up --build
```

## Run with Go

```powershell
go run . -dir .\docs -addr :8080
```

## Flags

| Flag | Default | Description |
| --- | --- | --- |
| `-dir` | `.` | Directory containing markdown files |
| `-addr` | `:8080` | Listen address |
| `-edit` | `auto` | Enable write routes: `auto`, `true`, or `false` |

`-edit=auto` enables the editor, upload, and new-file routes when the listen
address is loopback (for example `127.0.0.1:8080` or `localhost:8080`), and
disables them otherwise. An empty host such as `:8080` binds all interfaces
and enables editing with a warning; pass `-edit=false` to serve read-only.

## Routes

| Route | Description |
| --- | --- |
| `/` | Index of all `.md` files |
| `/<path>` | Renders `<path>.md` |
| `/__frag?p=<path>` | Rendered content fragment |
| `/__tree?route=<path>` | Sidebar tree fragment |
| `/__search?q=<term>` | JSON search results |
| `/ws` | WebSocket used for live reload |
| `/__edit?p=<path>` | Editor for a document |
| `/__render` | POST markdown, returns HTML preview |
| `/__save?p=<path>` | POST markdown, saves atomically |
| `/__new` | Create a new document |
| `/__upload` | Upload one or more `.md` files |

The `/__edit`, `/__render`, `/__save`, `/__new`, and `/__upload` routes
require editing to be enabled.

## Editing from the browser

Open any page and click **Edit**, or use the sidebar **New** and **Upload**
controls. The preview is rendered on the server with the same goldmark
pipeline as the published page. Saves are atomic (temp file plus rename) so
the watcher never reads a partially written file. The write routes can create
or overwrite any file under `-dir`, so keep the server on loopback or behind
authentication, and mount the directory read-write.

Rendering uses [goldmark](https://github.com/yuin/goldmark) with GitHub
Flavored Markdown. File watching uses
[fsnotify](https://github.com/fsnotify/fsnotify); reload push uses
[gorilla/websocket](https://github.com/gorilla/websocket).
