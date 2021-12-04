package main

import "log"

// Room struct for rooms support
type Room struct {
	hub     *Hub
	id      string
	usersID []string
	send    chan Message
}

func (r *Room) writePump() {
	for {
		message, ok := <-r.send
		if !ok {
			log.Println("Room was deleted")
		}
		for _, id := range r.usersID {
			message.Destination = id
			r.hub.broadcast <- message
		}
	}
}

// GetSendChan implementation of HubWriter.GetSendChan()
func (r *Room) GetSendChan() chan Message {
	return r.send
}

// GetID implementation of HubWriter.GetID()
func (r *Room) GetID() string {
	return r.id
}
