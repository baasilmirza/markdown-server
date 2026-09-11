package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// hub tracks connected browsers and broadcasts reload messages to them.
type hub struct {
	mu    sync.Mutex
	conns map[*websocket.Conn]struct{}
}

func newHub() *hub {
	return &hub{conns: make(map[*websocket.Conn]struct{})}
}

func (h *hub) add(c *websocket.Conn) {
	h.mu.Lock()
	h.conns[c] = struct{}{}
	h.mu.Unlock()
}

func (h *hub) remove(c *websocket.Conn) {
	h.mu.Lock()
	delete(h.conns, c)
	h.mu.Unlock()
}

func (h *hub) broadcast(msg string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.conns {
		if err := c.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			c.Close()
			delete(h.conns, c)
		}
	}
}

func (h *hub) handleWS(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("websocket upgrade:", err)
		return
	}
	h.add(c)
	defer func() {
		h.remove(c)
		c.Close()
	}()
	for {
		if _, _, err := c.ReadMessage(); err != nil {
			return
		}
	}
}
