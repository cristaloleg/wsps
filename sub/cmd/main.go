package main

import (
	"flag"

	"github.com/cristaloleg/wsps/queue"
	"github.com/cristaloleg/wsps/sub"
)

var (
	amqpURL string
	port    string
)

func init() {
	flag.StringVar(&amqpURL, "url", "amqp://127.0.0.1:5672", "AMQP url for the publisher")
	flag.StringVar(&port, "port", "3001", "Subscriber's port")
}

func main() {
	s := sub.Sub{}
	q, _ := queue.New(amqpURL)

	defer s.Close()

	s.Init(
		q,
		sub.WithRWBuf(1024),
		sub.WithPort(port),
	)
	s.Run()
}
