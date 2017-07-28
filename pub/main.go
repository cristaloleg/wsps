package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/cristaloleg/wsps/client"
	"github.com/cristaloleg/wsps/hub"
	"github.com/cristaloleg/wsps/message"
	"github.com/gorilla/websocket"
	"github.com/streadway/amqp"
)

type params struct {
	amqpURL string
	port    string
}

var param params

func init() {
	flag.StringVar(&param.amqpURL, "url", "amqp:///", "AMQP url for the publisher")
	flag.StringVar(&param.port, "port", "3030", "Publisher's port")
}

type pub struct {
	conn *amqp.Connection
	ch   *amqp.Channel

	hubs map[string]*hub.Hub

	messages chan message.Message
	done     chan struct{}

	upgrader websocket.Upgrader
}

func (p *pub) Init() {
	p.hubs = make(map[string]*hub.Hub, 16)

	p.messages = make(chan message.Message, 1024)
	p.done = make(chan struct{}, 1)

	p.upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}

	// conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	var err error
	p.conn, err = amqp.Dial(param.amqpURL + ":" + param.port)
	failOnError(err, "Failed to connect to RabbitMQ")

	p.ch, err = p.conn.Channel()
	failOnError(err, "Failed to open a channel")
}

func (p *pub) Close() {
	p.done <- struct{}{}

	if err := p.ch.Close(); err != nil {
		log.Println("cannot close channel")
	}

	if err := p.conn.Close(); err != nil {
		log.Println("cannot close connection")
	}

	for _, h := range p.hubs {
		h.Close()
	}
	p.hubs = nil

	close(p.messages)
	close(p.done)
}

func (p *pub) newQueue(name string) error {
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

func (p *pub) Listen() {
	go func() {
		for {
			select {
			case m := <-p.messages:
				p.publish(m.Topic, m.Raw())

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

func (p *pub) wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := p.upgrader.Upgrade(w, r, nil)
	failOnError(err, "Failed to upgrade a message")

	var msg message.Header
	if err := ws.ReadJSON(&msg); err != nil {
		err = ws.WriteMessage(websocket.TextMessage, []byte("incorrect header"))
		log.Println(err)
		return
	}

	err = p.register(msg.Topic, ws)
	if err != nil {
		//
	}
}

func (p *pub) register(name string, conn *websocket.Conn) error {
	h, ok := p.hubs[name]
	if !ok {
		if err := p.newQueue(name); err != nil {
			log.Println("cannot create queue")
			conn.Close()
			return err
		}

		h := hub.NewPub(p.messages)
		p.hubs[name] = h
	}
	c := client.New(conn)
	h.AddClient(c)
	return nil
}

func main() {
	var p pub

	defer p.Close()
	p.Init()
	p.Listen()
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
