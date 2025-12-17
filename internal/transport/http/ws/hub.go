package ws

import (
	"encoding/json"
	"sync"
)

type Hub struct {
	mu 		sync.RWMutex
	clients	map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
	}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.clients, c)
}

func (h *Hub) Broadcast(event any) {
	data, err := json.Marshal(event)
	if err != nil{
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients {
		select {
		case c.send <- data:
		default:
			
		}
	}
}