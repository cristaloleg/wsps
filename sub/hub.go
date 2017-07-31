package sub

import (
	"sync"

	"github.com/cristaloleg/wsps/client"
	"github.com/cristaloleg/wsps/hub"
	"github.com/cristaloleg/wsps/message"
)

var h hub.Huber = (*Hub)(nil)

// Hub ...
type Hub struct {
	mu        sync.RWMutex
	name      string
	clients   map[*client.Client]struct{}
	toClients <-chan message.Message
	done      chan struct{}
}

// NewHub ...
func NewHub(name string, ch <-chan message.Message) *Hub {
	h := &Hub{
		name:      name,
		clients:   make(map[*client.Client]struct{}),
		toClients: ch,
	}
	go h.listen()
	return h
}

func (h *Hub) listen() {
	for {
		select {
		case msg := <-h.toClients:
			defer h.mu.RUnlock()
			h.mu.RLock()
			for c := range h.clients {
				c.Send(msg)
			}

		case <-h.done:
			return
		}
	}
}

// AddClient ...
func (h *Hub) AddClient(c *client.Client) {
	defer h.mu.Unlock()
	h.mu.Lock()
	h.clients[c] = struct{}{}
}

// DelClient ...
func (h *Hub) DelClient(c *client.Client) {
	defer h.mu.Unlock()
	h.mu.Lock()
	delete(h.clients, c)
}

// Close ...
func (h *Hub) Close() error {
	h.done <- struct{}{}
	for c := range h.clients {
		c.Close()
	}
	return nil
}
