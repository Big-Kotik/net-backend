package main

import (
	"log"
	"sync"
)

type void struct {
}

type hub struct {
	mx sync.RWMutex

	hubID string

	clients map[HubWriter]void

	writers map[string]HubWriter

	broadcast chan ClientMessage
}

var (
	once        sync.Once
	hubInstance hub
)

// Hub interface for singleton hub struct
type Hub interface {
	GetID() string
	SendMessage(message ClientMessage)
	Register(client HubWriter)
	Unregister(client HubWriter)
	ContainsID(id string) bool
}

func (h *hub) GetID() string {
	return h.hubID
}

func (h *hub) SendMessage(message ClientMessage) {
	h.broadcast <- message
}

func (h *hub) Register(client HubWriter) {
	h.mx.Lock()
	defer h.mx.Unlock()
	h.clients[client] = void{}
	h.writers[client.GetID()] = client
}

func (h *hub) Unregister(client HubWriter) {
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
		message := <-h.broadcast
		if client, err := h.writers[message.Destination]; err {
			select {
			case client.GetSendChan() <- message:
			default:
				close(client.GetSendChan())
				delete(h.clients, client)
				delete(h.writers, message.Destination)
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
			hubID:     getID(),
			broadcast: make(chan ClientMessage),
			clients:   make(map[HubWriter]void),
			writers:   make(map[string]HubWriter),
		}
		go hubInstance.run()
	})
	return &hubInstance
}
