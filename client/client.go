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
	inCh        chan message.Body
	done        chan struct{}
	isListening bool
}

// New ...
func New(connection *websocket.Conn) *Client {
	c := &Client{
		conn: connection,
		inCh: make(chan message.Body, 1024),
		done: make(chan struct{}),
	}
	c.conn.SetReadLimit(defaultReadLimit)
	return c
}

// Close ...
func (c *Client) Close() {
	c.done <- struct{}{}
	close(c.inCh)

	if err := c.conn.Close(); err != nil {
		log.Println(err)
	}
}

// Write ...
func (c *Client) Write(m message.Body) {
	c.inCh <- m
}

// WriteBytes ...
func (c *Client) WriteBytes(m message.Body) {
	c.inCh <- m
}

// Listen ...
func (c *Client) Listen() {
	if c.isListening {
		return
	}
	c.isListening = true

	for {
		select {
		case m := <-c.inCh:
			if err := c.conn.WriteJSON(m); err != nil {
				log.Println(err)
			}

		case <-c.done:
			close(c.done)
			return
		}
	}
}

// SendTo ...
func (c *Client) SendTo(ch chan<- message.Message) {
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
