package pub

import (
	"sync"

	"github.com/cristaloleg/wsps/client"
	"github.com/cristaloleg/wsps/hub"
	"github.com/cristaloleg/wsps/message"
)

var h hub.Huber = (*Hub)(nil)

// Hub ...
type Hub struct {
	mu      sync.RWMutex
	name    string
	clients map[*client.Client]struct{}
	toAmqp  chan<- message.Message
	done    chan struct{}
}

// NewHub ...
func NewHub(name string, ch chan<- message.Message) *Hub {
	h := &Hub{
		name:    name,
		clients: make(map[*client.Client]struct{}),
		toAmqp:  ch,
	}
	return h
}

// AddClient ...
func (h *Hub) AddClient(c *client.Client) {
	c.PushTo(h.toAmqp)

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
