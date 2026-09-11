# Day 2

Second entry. Sibling of [Day 1](day-1.md).

## Observations

The recursive watcher emits one event per write. Atomic saves via temp file
plus rename produce a rename event, which the incremental reload treats as a
content change.

## Snippet

```json
{ "type": "reload", "path": "notes/2026/day-2.md", "op": "write" }
```

Back to the [index](../../README.md).
