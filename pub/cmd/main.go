package main

import (
	"flag"

	"github.com/cristaloleg/wsps/pub"
	"github.com/cristaloleg/wsps/queue"
)

var (
	amqpURL string
	port    string
)

func init() {
	flag.StringVar(&amqpURL, "url", "amqp:///", "AMQP url for the publisher")
	flag.StringVar(&port, "port", "3030", "Publisher's port")
}

func main() {
	p := pub.Pub{}
	q, _ := queue.New(amqpURL, port)

	defer p.Close()

	p.Init(
		q,
		pub.WithRWBuf(1024),
	)
	p.Run()
}
