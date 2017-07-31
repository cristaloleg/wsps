package client_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/gorilla/websocket"

	"github.com/cristaloleg/wsps/client"
	"github.com/cristaloleg/wsps/message"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

func TestClientRead(t *testing.T) {
	msg := message.Message{Topic: "interview"}

	srv := httptest.NewServer(wrapMockReadHandler(msg))
	u, _ := url.Parse(srv.URL)
	u.Scheme = "ws"

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("cannot make websocket connection: %v", err)
	}

	c := client.New(conn)
	if c == nil {
		t.Error("cannot instantiate Client")
	}

	go c.Listen()

	raw, err := msg.Marshal()
	if err != nil {
		t.Error("cannot marshal Message")
	}

	_, p, err := conn.ReadMessage()
	if err != nil {
		t.Errorf("cannot read message: %v", err)
	}

	if !reflect.DeepEqual(p, raw) {
		t.Error("message aren't equal")
	}

	c.Close()
	srv.Close()
}

func TestClientWrite(t *testing.T) {
	msg := message.Message{Topic: "interview"}

	srv := httptest.NewServer(wrapMockWriteHandler(msg))
	u, _ := url.Parse(srv.URL)
	u.Scheme = "ws"

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("cannot make websocket connection: %v", err)
	}

	c := client.New(conn)
	if c == nil {
		t.Error("cannot instantiate Client")
	}

	ch := make(chan message.Message, 1)
	c.PushTo(ch)

	c.Close()
	srv.Close()
}

func wrapMockReadHandler(msg message.Message) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot upgrade: %v", err), http.StatusInternalServerError)
		}

		raw, err := msg.Marshal()
		if err != nil {
			panic("cannot marshal Message")
		}

		conn.WriteMessage(websocket.TextMessage, raw)
	}
}

func wrapMockWriteHandler(msg message.Message) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("cannot upgrade: %v", err), http.StatusInternalServerError)
		}

		raw, err := msg.Marshal()
		conn.WriteMessage(websocket.TextMessage, raw)
	}
}
