package message

import (
	"bytes"
	"encoding/json"
	"time"
)

// Message between services
type Message struct {
	Type   string    `json:"type"`
	Token  string    `json:"token"`
	Topic  string    `json:"topic"`
	Create time.Time `json:"created,omitempty"`
	Author string    `json:"author"`
	Body   []byte    `json:"body"`
}

// Marshal to JSON
func (m *Message) Marshal() ([]byte, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal from JSON
func (m *Message) Unmarshal(raw []byte) error {
	buf := bytes.NewBuffer(raw)
	if err := json.NewDecoder(buf).Decode(&m); err != nil {
		return err
	}
	return nil
}
