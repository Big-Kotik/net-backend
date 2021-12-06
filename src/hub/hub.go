package hub

import (
	"log"
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

	writers map[string]Client

	broadcast chan msg.ClientMessage
}

var (
	once        sync.Once
	hubInstance hub
)

// Hub interface for singleton hub struct
type Hub interface {
	GetID() string
	SendMessage(message msg.ClientMessage)
	Register(client Client)
	Unregister(client Client)
	ContainsID(id string) bool
}

func (h *hub) GetID() string {
	return h.hubID
}

func (h *hub) SendMessage(message msg.ClientMessage) {
	h.broadcast <- message
}

func (h *hub) Register(client Client) {
	h.mx.Lock()
	defer h.mx.Unlock()
	h.clients[client] = void{}
	h.writers[client.GetID()] = client
}

func (h *hub) Unregister(client Client) {
	h.mx.Lock()
	defer h.mx.Unlock()
	delete(h.clients, client)
	delete(h.writers, client.GetID())
	close(client.GetSendChan())
}

func (h *hub) ContainsID(id string) bool {
	h.mx.RLock()
	defer h.mx.RUnlock()
	_, ok := h.writers[id]
	return ok
}

func (h *hub) run() {
	for {
		clientMessage := <-h.broadcast
		if client, err := h.writers[clientMessage.Destination]; err {
			select {
			case client.GetSendChan() <- clientMessage:
			default:
				close(client.GetSendChan())
				delete(h.clients, client)
				delete(h.writers, clientMessage.Destination)
			}
		} else {
			log.Println("No such channel")
		}
	}
}

// GetHub return Hub object
func GetHub() Hub {
	once.Do(func() {
		hubInstance = hub{
			hubID:     security.GetID(),
			broadcast: make(chan msg.ClientMessage),
			clients:   make(map[Client]void),
			writers:   make(map[string]Client),
		}
		go hubInstance.run()
	})
	return &hubInstance
}
