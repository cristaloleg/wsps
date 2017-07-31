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
	flag.StringVar(&amqpURL, "url", "amqp:///", "AMQP url for the subscriber")
	flag.StringVar(&port, "port", "3031", "Subscriber's port")
}

func main() {
	s := sub.Sub{}
	q, _ := queue.New(amqpURL, port)

	defer s.Close()

	s.Init(q,
		sub.WithRWBuf(1024),
	)
	s.Run()
}
