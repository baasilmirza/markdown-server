# FAQ

## Does it support nested directories?

Yes. The sidebar is a **recursive** tree and every subdirectory is watched.

## How does search work?

Search is server-side. The query runs against a catalog of every document and
returns ranked matches with a short snippet. Try searching for `kangaroo`.

## Why did my page not reload?

Incremental reload only swaps the content fragment for the document you are
viewing. If you changed a different file, the sidebar updates but your page
stays put.

## Can I edit files in the browser?

Yes, when the `-edit` flag is enabled. Use the **editor**, **New**, and
**Upload** controls. Files are written atomically so the watcher never reads a
half-written file.

## Is it safe to expose publicly?

No. The write routes let any reachable client create or overwrite files inside
`-dir`. Keep it on loopback or behind auth.

## Keyboard shortcuts

| Keys | Action |
| --- | --- |
| `/` | Focus search |
| `j` / `k` | Next / previous document |
| `Ctrl+S` | Save in editor |
| `Esc` | Close editor or search |
