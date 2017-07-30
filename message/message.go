package message

import (
	"bytes"
	"encoding/gob"
	"time"
)

// Message ...
type Message struct {
	Header Header
	Body   Body
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

// MarshalJSON ...
func (m *Message) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalJSON ...
func (m *Message) UnmarshalJSON(raw []byte) error {
	buf := bytes.NewBuffer(raw)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&m); err != nil {
		return err
	}
	return nil
}

// MarshalJSON ...
func (h *Header) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(h); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalJSON ...
func (h *Header) UnmarshalJSON(raw []byte) error {
	buf := bytes.NewBuffer(raw)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&h); err != nil {
		return err
	}
	return nil
}

// MarshalJSON ...
func (b *Body) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(b); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalJSON ...
func (b *Body) UnmarshalJSON(raw []byte) error {
	buf := bytes.NewBuffer(raw)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&b); err != nil {
		return err
	}
	return nil
}
