package client

import (
	"io"
	"log"

	"github.com/cristaloleg/wsps/message"
	"github.com/gorilla/websocket"
)

const defaultReadLimit = 1024

// Client ...
type Client struct {
	conn        *websocket.Conn
	ch          chan message.Message
	done        chan struct{}
	isListening bool
}

// New ...
func New(connection *websocket.Conn) *Client {
	c := &Client{
		conn: connection,
		ch:   make(chan message.Message, 1024),
		done: make(chan struct{}),
	}
	c.conn.SetReadLimit(defaultReadLimit)
	return c
}

// Close ...
func (c *Client) Close() {
	c.done <- struct{}{}
	close(c.ch)

	if err := c.conn.Close(); err != nil {
		log.Println(err)
	}
}

// Send ...
func (c *Client) Send(m message.Message) {
	c.ch <- m
}

// Listen ...
func (c *Client) Listen() {
	if c.isListening {
		return
	}
	c.isListening = true

	for {
		select {
		case m := <-c.ch:
			if err := c.conn.WriteJSON(m); err != nil {
				log.Println(err)
			}

		case <-c.done:
			close(c.done)
			return
		}
	}
}

// PushTo ...
func (c *Client) PushTo(ch chan<- message.Message) {
	for {
		select {
		default:
			var msg message.Message
			err := c.conn.ReadJSON(&msg)
			if err == io.ErrUnexpectedEOF || websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Println(err)
				break
			}

			ch <- msg

		case <-c.done:
			close(c.done)
			return
		}
	}
}
