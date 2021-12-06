package msg

// NodeMessageType show node message type
type NodeMessageType uint

const (
	// NewConnection is type for setup connection
	NewConnection NodeMessageType = iota
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
	Type        NodeMessageType `json:"type"`
	Body        string          `json:"body"`
}
