package hub

import "net-backend/src/msg"

// Client interface for writing msg by hub
type Client interface {
	GetSendChan() chan msg.ClientMessage
	GetID() string
}

// Node interface for writing msg by hub
type Node interface {
	GetSendChan() chan msg.NodeMessage
	GetID() string
}
