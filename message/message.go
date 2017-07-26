package message

import "time"

// Message represents a message transferred between microservices
type Message struct {
	Topic  string    `json:"topic"`
	Create time.Time `json:"created,omitempty"`
	Body   string    `json:"body"`
}
