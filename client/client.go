package client

import (
	"io"
	"log"

	"github.com/cristaloleg/wsps/message"
	"github.com/gorilla/websocket"
)

const defaultReadLimit = 1024

// Client is a websocket connection to the client
type Client struct {
	conn        *websocket.Conn
	ch          chan message.Message
	done        chan struct{}
	isListening bool
}

// New create a new client from websocket connection
func New(connection *websocket.Conn) *Client {
	c := &Client{
		conn: connection,
		ch:   make(chan message.Message, 1024),
		done: make(chan struct{}),
	}
	c.conn.SetReadLimit(defaultReadLimit)
	return c
}

// Close stops client from receiving/sending messages
func (c *Client) Close() error {
	c.done <- struct{}{}
	close(c.ch)

	if err := c.conn.Close(); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// Send sends a message to the client. Might be blocked
func (c *Client) Send(m message.Message) {
	c.ch <- m
}

// Listen listens to the incomming messages from the channel
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

// PushTo sends messages from the client to the given channel
func (c *Client) PushTo(ch chan<- message.Message) {
	go func() {
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
	}()
}
