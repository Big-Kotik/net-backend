package hub

import (
	"net-backend/src/msg"
	"net-backend/src/security"
	"sync"
)

type void struct {
}

type hub struct {
	mx sync.RWMutex

	hubID string

	clients map[Client]void

	nodes map[Node]void

	idToClient map[string]Client
	idToNode   map[string]Node

	clientMessageBroadcast chan msg.ClientMessage

	nodeMessageBroadcast chan msg.NodeMessage
}

var (
	once        sync.Once
	hubInstance hub
)

// Hub interface for singleton hub struct
type Hub interface {
	GetID() string
	GetNodeIDs() []string
	GetClientIDs() []string
	SendMessage(message msg.ClientMessage)
	RegisterClient(client Client)
	UnregisterClient(client Client)
	RegisterNode(node Node)
	UnregisterNode(node Node)
	ContainsID(id string) bool
	BroadcastNodeMessage(message msg.NodeMessage, noBroadcastNodeID map[string]struct{})
}

func (h *hub) GetID() string {
	return h.hubID
}

func (h *hub) SendMessage(message msg.ClientMessage) {
	h.clientMessageBroadcast <- message
}

func (h *hub) RegisterClient(client Client) {
	h.mx.Lock()
	defer h.mx.Unlock()
	h.clients[client] = void{}
	h.idToClient[client.GetID()] = client
}

func (h *hub) UnregisterClient(client Client) {
	h.mx.Lock()
	defer h.mx.Unlock()
	delete(h.clients, client)
	delete(h.idToClient, client.GetID())
	close(client.GetSendChan())
}

func (h *hub) RegisterNode(node Node) {
	h.mx.Lock()
	defer h.mx.Unlock()
	h.nodes[node] = void{}
	h.idToNode[node.GetID()] = node
}

func (h *hub) UnregisterNode(node Node) {
	h.mx.Lock()
	defer h.mx.Unlock()
	delete(h.nodes, node)
	delete(h.idToNode, node.GetID())
	close(node.GetSendChan())
}

func (h *hub) ContainsID(id string) bool {
	h.mx.RLock()
	defer h.mx.RUnlock()
	_, ok := h.idToClient[id]
	return ok
}

func (h *hub) GetNodeIDs() []string {
	h.mx.RLock()
	defer h.mx.RUnlock()
	ans := make([]string, 0, len(h.nodes))
	for key := range h.nodes {
		ans = append(ans, key.GetID())
	}
	return ans
}

func (h *hub) GetClientIDs() []string {
	h.mx.RLock()
	defer h.mx.RUnlock()
	ans := make([]string, 0, len(h.clients))
	for key := range h.clients {
		ans = append(ans, key.GetID())
	}
	return ans
}

func (h *hub) BroadcastNodeMessage(message msg.NodeMessage, noBroadcastNodeID map[string]struct{}) {
	for node := range h.nodes {
		if _, ok := noBroadcastNodeID[node.GetID()]; !ok {
			node.GetSendChan() <- message
		}
	}
}

func (h *hub) run() {
	for {
		clientMessage := <-h.clientMessageBroadcast
		if client, err := h.idToClient[clientMessage.Destination]; err {
			select {
			case client.GetSendChan() <- clientMessage:
			default:
				close(client.GetSendChan())
				delete(h.clients, client)
				delete(h.idToClient, clientMessage.Destination)
			}
		} else {
			message, err := msg.FromClientMessage(clientMessage)
			if err != nil {
				continue
			}
			h.BroadcastNodeMessage(message, map[string]struct{}{})
		}
	}
}

// GetHub return Hub object
func GetHub() Hub {
	once.Do(func() {
		hubInstance = hub{
			hubID:                  security.GetID(),
			clientMessageBroadcast: make(chan msg.ClientMessage),
			nodeMessageBroadcast:   make(chan msg.NodeMessage),
			clients:                make(map[Client]void),
			idToClient:             make(map[string]Client),
		}
		go hubInstance.run()
	})
	return &hubInstance
}
