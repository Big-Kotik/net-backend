package main

import (
	"log"
)

type Message struct {
	Destination string `json:"destination"`
	Source      string `json:"source"`
	Message     string `json:"message"`
}

type void struct {
}

type Hub struct {
	clients map[HubWriter]void

	writers map[string]HubWriter

	broadcast chan Message

	register chan HubWriter

	unregister chan HubWriter
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan HubWriter),
		unregister: make(chan HubWriter),
		clients:    make(map[HubWriter]void),
		writers:    make(map[string]HubWriter),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = void{}
			h.writers[client.GetId()] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.writers, client.GetId())
				close(*client.GetSendChan())
			}
		case message := <-h.broadcast:
			if client, err := h.writers[message.Destination]; err {
				select {
				case *client.GetSendChan() <- message:
				default:
					close(*client.GetSendChan())
					delete(h.clients, client)
					delete(h.writers, message.Destination)
				}
			} else {
				log.Println("No such channel")
			}
		}
	}
}
