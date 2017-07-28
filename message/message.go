package message

import (
	"bytes"
	"encoding/gob"
	"time"
)

// Message ...
type Message struct {
	Header
	Body
}

// Header ...
type Header struct {
	Type  string `json:"type"`
	Token string `json:"token"`
	Topic string `json:"topic"`
}

// Body ...
type Body struct {
	Create time.Time `json:"created,omitempty"`
	Author string    `json:"author"`
	Body   []byte    `json:"body"`
}

// Raw ...
func (m *Message) Raw() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// FromRaw ...
func FromRaw(raw []byte) (*Message, error) {
	buf := bytes.NewBuffer(raw)
	dec := gob.NewDecoder(buf)
	var data Message
	if err := dec.Decode(data); err != nil {
		return nil, err
	}
	return &data, nil

}
