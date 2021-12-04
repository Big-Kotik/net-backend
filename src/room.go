package main

import "log"

type Room struct {
	hub     *Hub
	id      string
	usersId []string
	send    chan Message
}

func (r *Room) writePump() {
	for {
		select {
		case message, ok := <-r.send:
			if !ok {
				log.Println("Room was deleted")
			}
			for _, id := range r.usersId {
				message.Destination = id
				r.hub.broadcast <- message
			}
		}
	}
}

func (r *Room) GetSendChan() *chan Message {
	return &r.send
}

func (r *Room) GetId() string {
	return r.id
}
