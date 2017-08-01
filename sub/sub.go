package sub

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

// Sub ...
type Sub struct {
	queue queue.Queue

	mu   sync.RWMutex
	hubs map[string]*Hub

	done chan struct{}

	upgrader websocket.Upgrader

	port string
}

type options struct {
	bufSize int
	port    string
}

// Option ...
type Option func(*options)

// WithRWBuf ...
func WithRWBuf(size int) Option {
	return func(o *options) {
		o.bufSize = size
	}
}

// WithPort ...
func WithPort(port string) Option {
	return func(o *options) {
		o.port = port
	}
}

// Init ...
func (s *Sub) Init(queue queue.Queue, opts ...Option) {
	var opt options
	for _, op := range opts {
		op(&opt)
	}

	s.queue = queue
	s.hubs = make(map[string]*Hub, 16)
	s.done = make(chan struct{})
	s.upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	s.port = opt.port
}

// Close ...
func (s *Sub) Close() {
	s.done <- struct{}{}
	s.queue.Close()

	for _, h := range s.hubs {
		h.Close()
	}
	s.hubs = nil

	close(s.done)
}

// Run ...
func (s *Sub) Run() {
	go func() {
		for {
			select {
			case <-s.done:
				return
			}
		}
	}()

	http.HandleFunc("/ws", s.wsHandler)
	if err := http.ListenAndServe(":"+s.port, nil); err != nil {
		failOnError(err, "server error")
	}
}

func (s *Sub) wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade a message", err)
		return
	}

	var msg message.Message
	if err := ws.ReadJSON(&msg); err != nil {
		log.Println(err)
		err = ws.WriteMessage(websocket.TextMessage, []byte("incorrect message format"))
		if err != nil {
			log.Println(err)
		}
		return
	}

	if err = s.register(msg.Topic, ws); err != nil {
		log.Println(err)
		return
	}
}

func (s *Sub) register(name string, conn *websocket.Conn) error {
	s.mu.RLock()
	h, ok := s.hubs[name]
	s.mu.RUnlock()

	if !ok {
		ch, err := s.queue.Get(name)
		if err != nil {
			log.Println("topic doesn't exist")
			return err
		}

		h = NewHub(name, ch)
		s.mu.Lock()
		s.hubs[name] = h
		s.mu.Unlock()
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
