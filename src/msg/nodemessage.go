package msg

import "encoding/json"

// NodeMessageType show node message type
type NodeMessageType uint

const (
	// GetNodeID is type for setup connection
	GetNodeID NodeMessageType = iota
	// ForwardClient is type for forwarding message between two node
	ForwardClient
	// GetClientsIDs is type for getting all client IDs which connected to this node
	GetClientsIDs
	// GetNodesAddress is type for getting all node IDs which connected to this node
	GetNodesAddress
)

// NodeMessage struct for msg which came from another node
type NodeMessage struct {
	Destination string          `json:"destination"`
	Source      string          `json:"source"`
	Body        string          `json:"body"`
	Type        NodeMessageType `json:"type"`
	NodeThrough uint32          `json:"nodeThrough"`
	IsAnswer    bool            `json:"answer"`
	IsError     bool            `json:"error"`
}

// SetType set type of message
func (m *NodeMessage) SetType(t NodeMessageType) {
	m.Type = t
}

// FromClientMessage wrap client message to forward message
func FromClientMessage(message ClientMessage) (NodeMessage, error) {
	bytes, err := json.Marshal(message)
	if err != nil {
		return NodeMessage{}, err
	}
	return NodeMessage{
		Destination: message.Destination,
		Source:      message.Source,
		Type:        ForwardClient,
		NodeThrough: 255,
		Body:        string(bytes),
	}, nil
}
