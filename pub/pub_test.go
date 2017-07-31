package pub_test

import (
	"testing"
)

func TestPub(t *testing.T) {
	// var q queue.FakeQueue
	// var p pub.Pub
	// p.Init(q)

	// // p.Run()

	// msg := message.Message{Topic: "interview"}
	// q.Publish(msg)

	// //create fake client
	// srv := httptest.NewServer(wrapMockReadHandler(msg))
	// u, _ := url.Parse(srv.URL)
	// u.Scheme = "ws"
	// conn, _, _ := websocket.DefaultDialer.Dial(u.String(), nil)

	// conn.WriteJSON(msg)

	// ch, _ := q.Get("")
	// raw, _ := msg.Marshal()
	// if m := <-ch; !reflect.DeepEqual(m, raw) {
	// 	t.Error("messages must be equal")
	// }
}
