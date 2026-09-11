# Advanced Usage

This page sits one level deep so the sidebar has a nested branch to draw.

## Recursive watching

The watcher is recursive. Every subdirectory is added to the fsnotify watch
set, and new directories are picked up as they appear. That means a file like
`notes/2026/day-1.md` reloads the browser just like a top-level file.

## Navigation

Use the **prev** and **next** links at the bottom of each page, or the left and
right arrow keys, to move through the document tree in order.

## Theme

The **theme** toggle in the sidebar switches between light and dark. Your
choice is stored in `localStorage` and applied before first paint so there is
no flash of the wrong theme.

## Editor

Click **Edit** on any page to open the in-browser **editor**. The preview is
rendered on the server with the same goldmark pipeline, so what you see is what
the reader gets.

## Upload

The sidebar exposes **New** and **Upload** controls for creating markdown files
without leaving the browser.

## Deployment table

| Flag | Default | Notes |
| --- | --- | --- |
| `-dir` | `.` | Root directory to serve |
| `-addr` | `:8080` | Listen address |
| `-edit` | loopback | Enable write routes |

See the [API Reference](../reference/api.md) for route details.
