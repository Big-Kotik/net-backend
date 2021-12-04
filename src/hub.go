package main

import (
	"log"
)

type Message struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

type void struct {
}

type Hub struct {
	clients map[HubWriter]void

	rooms map[string]HubWriter

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
		rooms:      make(map[string]HubWriter),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = void{}
			h.rooms[client.GetId()] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.rooms, client.GetId())
				close(*client.GetSendChan())
			}
		case message := <-h.broadcast:
			if client, err := h.rooms[message.Id]; err {
				select {
				case *client.GetSendChan() <- []byte(message.Message):
				default:
					close(*client.GetSendChan())
					delete(h.clients, client)
					delete(h.rooms, message.Id)
				}
			} else {
				log.Println("No such channel")
			}
		}
	}
}
