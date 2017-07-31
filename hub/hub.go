package hub

import (
	"github.com/cristaloleg/wsps/client"
)

// Huber is an interface for Hubs
type Huber interface {
	AddClient(*client.Client)
	DelClient(*client.Client)
	Close() error
}
