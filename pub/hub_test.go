package pub_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/cristaloleg/wsps/client"
	"github.com/cristaloleg/wsps/message"
	"github.com/cristaloleg/wsps/pub"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

func TestHub(t *testing.T) {
	ch := make(chan message.Message, 1)
	h := pub.NewHub("", ch)
	msg := message.Message{Topic: "lol"}

	srv := httptest.NewServer(http.HandlerFunc(wrapMockReadHandler(msg)))
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
	h.AddClient(c)

	if len(ch) != 0 {
		t.Error("must be readed by client")
	}

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
