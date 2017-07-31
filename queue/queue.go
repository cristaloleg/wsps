package queue

import (
	"log"

	"github.com/cristaloleg/wsps/message"
	"github.com/streadway/amqp"
)

// Queue ...
type Queue interface {
	Create(string) error
	Publish(message.Message) error
	Get(string) (<-chan message.Message, error)
	Close() error
}

// New ...
func New(url, port string) (Queue, error) {
	conn, err := amqp.Dial(url + ":" + port)
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ", err)
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel", err)
		return nil, err
	}

	q := &queue{
		conn: conn,
		ch:   ch,
		done: make(chan struct{}),
	}
	return q, nil
}

type queue struct {
	conn       *amqp.Connection
	ch         *amqp.Channel
	done       chan struct{}
	numOfChans int
}

// Publish ...
func (q *queue) Publish(msg message.Message) error {
	raw, err := msg.Marshal()
	if err != nil {
		return err
	}

	return q.ch.Publish(
		"",        // exchange
		msg.Topic, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        raw,
		},
	)
}

// Close ...
func (q *queue) Close() error {
	for i := 0; i < q.numOfChans; i++ {
		q.done <- struct{}{}
	}

	if err := q.ch.Close(); err != nil {
		log.Println("cannot close channel")
		return err
	}

	if err := q.conn.Close(); err != nil {
		log.Println("cannot close connection")
		return err
	}

	return nil
}

// Create ...
func (q *queue) Create(name string) error {
	_, err := q.ch.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err == nil {
		q.numOfChans++
	}
	return err
}

// Get ...
func (q *queue) Get(name string) (<-chan message.Message, error) {
	delivery, err := q.ch.Consume(
		name,  // queue
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	ch := make(chan message.Message)

	go func(in <-chan amqp.Delivery) {
		for {
			select {
			case d := <-in:
				ch <- message.Message{Body: d.Body}

			case <-q.done:
				return
			}
		}
	}(delivery)

	return ch, err
}

var _ Queue = (*FakeQueue)(nil)

type FakeQueue struct {
	ch chan message.Message
}

func (f FakeQueue) Create(string) error {
	return nil
}

func (f FakeQueue) Publish(msg message.Message) error {
	f.ch <- msg
	return nil
}

func (f FakeQueue) Get(string) (<-chan message.Message, error) {
	return f.ch, nil
}

func (f FakeQueue) Close() error {
	return nil
}
