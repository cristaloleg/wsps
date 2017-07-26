package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"sync"

	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

type pub struct {
	mutex    sync.RWMutex
	clients  map[*websocket.Conn]struct{}
	messages chan message
	done     chan struct{}
	upgrader websocket.Upgrader
	conn     *amqp.Connection
	ch       *amqp.Channel
}

type message struct {
	Topic  string    `json:"topic"`
	Create time.Time `json:"created,omitempty"`
	Body   string    `json:"body"`
}

func (p *pub) init() {
	p.clients = make(map[*websocket.Conn]struct{}, 1024)
	p.messages = make(chan message)
	p.upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}
}

func (p *pub) dial() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	p.conn = conn
}

func (p *pub) close() {
	p.done <- struct{}{}

	if err := p.ch.Close(); err != nil {
		log.Println("cannot close channel")
	}

	if err := p.conn.Close(); err != nil {
		log.Println("cannot close connection")
	}

	close(p.messages)
	close(p.done)
}

func (p *pub) queue(name string) error {
	_, err := p.ch.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	return err
}

func (p *pub) publish(name string, message []byte) error {
	return p.ch.Publish(
		"",    // exchange
		name,  // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        message,
		},
	)
}

func (p *pub) listen() {
	for {
		select {
		case msg, ok := <-p.messages:
			if ok {
				if err := p.publish(msg.Topic, []byte(msg.Body)); err != nil {
					println(err)
				}
			}

		case <-p.done:
			return
		}
	}
}

func (p *pub) wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := p.upgrader.Upgrade(w, r, nil)
	failOnError(err, "Failed to upgrade a message")

	p.mutex.Lock()
	p.clients[ws] = struct{}{}
	p.mutex.Unlock()

	for {
		var msg message
		if err := ws.ReadJSON(&msg); err == io.ErrUnexpectedEOF {
			p.mutex.Lock()
			delete(p.clients, ws)
			p.mutex.Unlock()

			if err = ws.Close(); err != nil {
				log.Println("error on closing socket")
			}
		} else if err != nil {
			log.Println(err)
			continue
		}
		p.messages <- msg
	}
}

func (p *pub) queueHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[1:]

	if r.Method == "POST" {
		if err := p.queue(name); err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// if r.Method == "DELETE" {
	// 	// if err := p.queue(name); err != nil {
	// 	// 	w.WriteHeader(http.StatusBadRequest)
	// 	// }
	// 	w.WriteHeader(http.StatusOK)
	// 	return
	// }

	w.WriteHeader(http.StatusMethodNotAllowed)
}

func main() {
	var p pub

	p.dial()
	defer p.close()
	go p.listen()

	http.HandleFunc("/", p.queueHandler)
	http.HandleFunc("/ws", p.wsHandler)

	if err := http.ListenAndServe(":3000", nil); err != nil {
		failOnError(err, "server error")
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
