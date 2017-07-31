package pub

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/cristaloleg/wsps/client"
	"github.com/cristaloleg/wsps/message"
	"github.com/cristaloleg/wsps/queue"

	"github.com/gorilla/websocket"
)

type options struct {
	bufSize int
}

// Option ...
type Option func(*options)

// WithRWBuf ...
func WithRWBuf(size int) Option {
	return func(o *options) {
		o.bufSize = size
	}
}

// Pub ...
type Pub struct {
	queue queue.Queue

	mu   sync.RWMutex
	hubs map[string]*Hub

	messages chan message.Message
	done     chan struct{}

	upgrader websocket.Upgrader
}

// Init ...
func (p *Pub) Init(queue queue.Queue, opts ...Option) {
	var opt options
	for _, op := range opts {
		op(&opt)
	}

	p.queue = queue
	p.hubs = make(map[string]*Hub, 16)
	p.messages = make(chan message.Message, 1024)
	p.done = make(chan struct{})
	p.upgrader = websocket.Upgrader{
		ReadBufferSize:  opt.bufSize,
		WriteBufferSize: opt.bufSize,
	}
}

// Close ...
func (p *Pub) Close() {
	p.done <- struct{}{}
	p.queue.Close()

	for _, h := range p.hubs {
		h.Close()
	}
	p.hubs = nil

	close(p.messages)
	close(p.done)
}

// Run ...
func (p *Pub) Run() {
	go func() {
		for {
			select {
			case m := <-p.messages:
				if err := p.queue.Publish(m); err != nil {
					log.Println("error message, skipping")
				}

			case <-p.done:
				return
			}
		}
	}()

	http.HandleFunc("/ws", p.wsHandler)
	if err := http.ListenAndServe(":3000", nil); err != nil {
		failOnError(err, "server error")
	}
}

func (p *Pub) wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := p.upgrader.Upgrade(w, r, nil)
	failOnError(err, "Failed to upgrade a message")

	var msg message.Message
	if err := ws.ReadJSON(&msg); err != nil {
		err = ws.WriteMessage(websocket.TextMessage, []byte("incorrect header"))
		log.Println(err)
		return
	}

	if err = p.register(msg.Topic, ws); err != nil {
		log.Println(err)
		return
	}
}

func (p *Pub) register(name string, conn *websocket.Conn) error {
	p.mu.RLock()
	h, ok := p.hubs[name]
	p.mu.RUnlock()

	if !ok {
		if err := p.queue.Create(name); err != nil {
			log.Println("Cannot  queue", err)
			return err
		}

		h := NewHub(name, p.messages)
		p.mu.Lock()
		p.hubs[name] = h
		p.mu.Unlock()
	}
	c := client.New(conn)
	h.AddClient(c)
	return nil
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
