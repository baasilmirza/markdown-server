package main

import (
	"flag"
	"log"
	"net"
	"net/http"
)

func main() {
	dir := flag.String("dir", ".", "directory containing markdown files")
	addr := flag.String("addr", ":8080", "address to listen on")
	edit := flag.String("edit", "auto", "enable editor/upload routes: auto, true, or false")
	flag.Parse()

	enableEdit, warn := editEnabled(*addr, *edit)
	if warn != "" {
		log.Print("warning: ", warn)
	}
	if !enableEdit {
		log.Print("write routes disabled (editor, upload, new file)")
	}

	h := newHub()
	s := &server{dir: *dir, edit: enableEdit}

	go watch(*dir, h)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", h.handleWS)
	mux.HandleFunc("/__search", s.handleSearch)
	mux.HandleFunc("/__frag", s.handleFrag)
	mux.HandleFunc("/__tree", s.handleTree)
	mux.HandleFunc("/__edit", s.handleEdit)
	mux.HandleFunc("/__render", s.handleRender)
	mux.HandleFunc("/__save", s.handleSave)
	mux.HandleFunc("/__new", s.handleNew)
	mux.HandleFunc("/__upload", s.handleUpload)
	mux.HandleFunc("/", s.handle)

	log.Printf("serving %s at http://localhost%s", *dir, *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}

// editEnabled resolves the -edit flag. "auto" enables write routes for
// loopback addresses and disables them otherwise.
func editEnabled(addr, mode string) (bool, string) {
	switch mode {
	case "true", "1", "yes", "on":
		return true, ""
	case "false", "0", "no", "off":
		return false, ""
	}
	host := addr
	if h, _, err := net.SplitHostPort(addr); err == nil {
		host = h
	}
	if host == "" {
		return true, "write routes enabled and reachable on all interfaces; set -edit=false to disable"
	}
	if host == "localhost" {
		return true, ""
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return true, ""
	}
	return false, ""
}
