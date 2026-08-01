package hub

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	Clients map[*websocket.Conn]bool
	mu      sync.Mutex
}

func NewHub() *Hub {
	return &Hub{Clients: make(map[*websocket.Conn]bool)}
}

func (h *Hub) Register(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Clients[conn] = true
}

func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.Clients, conn)
}

func (h *Hub) Broadcast(message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn := range h.Clients {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("websocket write message failed: %v", err)
		}
	}
}
