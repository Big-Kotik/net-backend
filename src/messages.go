package main

// ClientMessage struct for message which came from client
type ClientMessage struct {
	Destination string `json:"destination"`
	Source      string `json:"source"`
	Message     string `json:"message"`
}

type NodeMessageType uint

const (
	NewConnection NodeMessageType = iota
	ForwardClientMessage
	GetClientsIDs
	GetKnownRoutersAddress
)

// NodeMessage struct for message which came from another node
type NodeMessage struct {
	Destination string          `json:"destination"`
	Source      string          `json:"source"`
	Type        NodeMessageType `json:"type"`
	Body        string          `json:"body"`
}
