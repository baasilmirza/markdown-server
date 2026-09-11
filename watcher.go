package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// watch monitors dir recursively and tells connected browsers to reload when a
// markdown file changes.
func watch(dir string, h *hub) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("watcher:", err)
		return
	}
	defer w.Close()

	if err := addRecursive(w, dir); err != nil {
		log.Println("watch add:", err)
		return
	}
	log.Printf("watching %s for changes", dir)

	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create != 0 {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					if err := addRecursive(w, event.Name); err != nil {
						log.Println("watch add:", err)
					}
				}
			}
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Remove) == 0 {
				continue
			}
			if !strings.EqualFold(filepath.Ext(event.Name), ".md") {
				continue
			}
			rel, err := filepath.Rel(dir, event.Name)
			if err != nil {
				continue
			}
			op := opName(event.Op)
			log.Printf("changed: %s (%s)", event.Name, op)
			h.broadcast(changeMsg{
				Type: "reload",
				Path: filepath.ToSlash(rel),
				Op:   op,
				Tree: op != "write",
			})
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Println("watch error:", err)
		}
	}
}

// opName returns the first matching fsnotify operation as a string.
func opName(op fsnotify.Op) string {
	switch {
	case op&fsnotify.Create != 0:
		return "create"
	case op&fsnotify.Remove != 0:
		return "remove"
	case op&fsnotify.Rename != 0:
		return "rename"
	default:
		return "write"
	}
}

// addRecursive adds dir and every subdirectory to the watcher.
func addRecursive(w *fsnotify.Watcher, dir string) error {
	return filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if p != dir && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		if err := w.Add(p); err != nil {
			return err
		}
		return nil
	})
}
