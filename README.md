# markdown-server

Serves `.md` files rendered to HTML and live-reloads the browser over a
WebSocket whenever a file changes.

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

## Routes

| Route | Description |
| --- | --- |
| `/` | Index of `.md` files in the directory |
| `/<name>` | Renders `<name>.md` |
| `/ws` | WebSocket used for live reload |

Rendering uses [goldmark](https://github.com/yuin/goldmark) with GitHub
Flavored Markdown. File watching uses
[fsnotify](https://github.com/fsnotify/fsnotify); reload push uses
[gorilla/websocket](https://github.com/gorilla/websocket).
