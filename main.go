package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	dir := flag.String("dir", ".", "directory containing markdown files")
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	h := newHub()
	s := &server{dir: *dir}

	go watch(*dir, h)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", h.handleWS)
	mux.HandleFunc("/__search", s.handleSearch)
	mux.HandleFunc("/", s.handle)

	log.Printf("serving %s at http://localhost%s", *dir, *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}
