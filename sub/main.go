// func processMessages() {
// 	messages, err := ch.Consume(
// 		q.Name, // queue
// 		"",     // consumer
// 		true,   // auto-ack
// 		false,  // exclusive
// 		false,  // no-local
// 		false,  // no-wait
// 		nil,    // args
// 	)
// 	failOnError(err, "Failed to register a consumer")

// 	for {
// 		// msg := <-messages
// 		for msg := range messages {
// 				if err := client.WriteMessage(websocket.TextMessage, msg.Body); err != nil {
// 					log.Println(err)
// 				}
// 			}
// 		}
// 	}
// }

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

type sub struct {
	conn *amqp.Connection
	ch   *amqp.Channel

	hubs map[string]*hub.Hub

	messages chan message.Message
	done     chan struct{}

	upgrader websocket.Upgrader
}

type params struct {
	amqpURL string
	port    string
}

var param params

func init() {
	flag.StringVar(&param.amqpURL, "url", "amqp:///", "AMQP url for the subscriber")
	flag.StringVar(&param.port, "port", "3031", "Subscriber's port")
}

func (s *sub) Init() {
	s.hubs = make(map[string]*hub.Hub, 16)

	// p.messages = make(chan message.Message, 1024)
	s.done = make(chan struct{}, 1)

	s.upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
	}

	var err error
	s.conn, err = amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")

	s.ch, err = s.conn.Channel()
	failOnError(err, "Failed to open a channel")
}

func (s *sub) Close() {
	s.done <- struct{}{}

	if err := s.ch.Close(); err != nil {
		log.Println("cannot close channel")
	}

	if err := s.conn.Close(); err != nil {
		log.Println("cannot close connection")
	}

	for _, h := range s.hubs {
		h.Close()
	}
	s.hubs = nil

	close(s.messages)
	close(s.done)
}

func (s *sub) subscribe(name string, client *websocket.Conn) error {
	msgs, err := s.ch.Consume(
		"name", // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		failOnError(err, "Failed to register a consumer")
		return err
	}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	return nil
}

func (s *sub) Listen() {
	go func() {
		for {
			select {
			case <-s.done:
				return
			}
		}
	}()

	http.HandleFunc("/ws", s.wsHandler)
	if err := http.ListenAndServe(":3001", nil); err != nil {
		failOnError(err, "server error")
	}
}

func (s *sub) getQueue(name string) *amqp.Queue {
	return nil
}

func (s *sub) wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	failOnError(err, "Failed to upgrade a message")

	var msg message.Header
	if err := ws.ReadJSON(&msg); err != nil {
		err = ws.WriteMessage(websocket.TextMessage, []byte("incorrect header"))
		log.Println(err)
		return
	}

	err = s.register(msg.Topic, ws)
	if err != nil {
		//
	}
}

// ws, err := s.upgrader.Upgrade(w, r, nil)
// failOnError(err, "Failed to upgrade a message")

// c := client.NewClient(ws)
// // go c.ListenFrom(s.messages)

// name := c.ListeningTo()
// h, ok := s.hubs[name]
// if !ok {
// 	q := s.getQueue(name)
// 	// h = hub.New(q.)
// 	s.hubs[name] = h
// }
// h.AddClient(c)

// for {
// 	var msg string
// 	if err := ws.ReadJSON(&msg); err == io.ErrUnexpectedEOF {
// 		if err = ws.Close(); err != nil {
// 			log.Println("error on closing socket")
// 		}
// 	} else if err != nil {
// 		log.Println(err)
// 		continue
// 	}

// 	if !s.queueExists(msg) {
// 		if err := ws.WriteMessage(websocket.TextMessage, []byte("no channel like this")); err != nil {
// 			log.Println(err)
// 		}
// 	}
// 	s.subscribe(msg, ws)
// }

func (s *sub) getChan(name string) <-chan amqp.Delivery {
	messages, err := s.ch.Consume(
		name,  // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	failOnError(err, "Failed to register a consumer")
	return messages
}

func (s *sub) register(name string, conn *websocket.Conn) error {
	h, ok := s.hubs[name]
	if !ok {
		// ch := s.getChan(name)

		h := hub.NewSub(nil)
		s.hubs[name] = h
	}
	c := client.New(conn)
	h.AddClient(c)
	return nil
}

func main() {
	var s sub

	defer s.Close()
	s.Init()
	s.Listen()
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}
