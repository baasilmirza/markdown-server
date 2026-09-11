package main

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

// watch monitors dir and tells connected browsers to reload when a markdown
// file changes.
func watch(dir string, h *hub) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println("watcher:", err)
		return
	}
	defer w.Close()

	if err := w.Add(dir); err != nil {
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
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Remove) == 0 {
				continue
			}
			if strings.EqualFold(filepath.Ext(event.Name), ".md") {
				log.Printf("changed: %s", event.Name)
				h.broadcast("reload")
			}
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			log.Println("watch error:", err)
		}
	}
}
