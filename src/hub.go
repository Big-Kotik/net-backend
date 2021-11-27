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
	clients map[*Client]void

	rooms map[string]*Client

	broadcast chan Message

	register chan *Client

	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]void),
		rooms:      make(map[string]*Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = void{}
			h.rooms[client.id] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				delete(h.rooms, client.id)
				close(client.send)
			}
		case message := <-h.broadcast:
			if client, err := h.rooms[message.Id]; err {
				select {
				case client.send <- []byte(message.Message):
				default:
					close(client.send)
					delete(h.clients, client)
					h.rooms[message.Id] = nil
				}
			} else {
				log.Println("No such channel")
			}
		}
	}
}
