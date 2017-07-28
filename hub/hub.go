package hub

import (
	"sync"

	"github.com/cristaloleg/wsps/client"
	"github.com/cristaloleg/wsps/message"
	"github.com/streadway/amqp"
)

// Hub ...
type Hub struct {
	sync.RWMutex
	clients  map[*client.Client]struct{}
	toAmqp   chan<- message.Message
	toUser   <-chan message.Message
	delivery <-chan amqp.Delivery
	done     chan struct{}
	isPub    bool
}

// NewPub ...
func NewPub(ch chan<- message.Message) *Hub {
	h := &Hub{
		clients: make(map[*client.Client]struct{}),
		toAmqp:  ch,
		isPub:   true,
	}
	return h
}

// NewSub ...
func NewSub(ch <-chan message.Message) *Hub {
	h := &Hub{
		clients: make(map[*client.Client]struct{}),
		toUser:  ch,
		isPub:   false,
	}
	go h.listen()
	return h
}

func (h *Hub) listen() {
	for {
		select {
		case m := <-h.delivery:
			defer h.RUnlock()
			h.RLock()
			for c := range h.clients {
				c.Write(message.FromRaw(m.Body))
			}

		case m := <-h.toUser:
			defer h.RUnlock()
			h.RLock()
			for c := range h.clients {
				c.Write(m.Body)
			}

		case <-h.done:
			return
		}
	}
}

// AddClient ...
func (h *Hub) AddClient(c *client.Client) {
	if h.isPub {
		go c.SendTo(h.toAmqp)
	} else {
		go c.Listen()
	}

	defer h.Unlock()
	h.Lock()
	h.clients[c] = struct{}{}
}

// DelClient ...
func (h *Hub) DelClient(c *client.Client) {
	defer h.Unlock()
	h.Lock()
	delete(h.clients, c)
}

// Close ...
func (h *Hub) Close() {
	h.done <- struct{}{}
	for c := range h.clients {
		c.Close()
	}
}
