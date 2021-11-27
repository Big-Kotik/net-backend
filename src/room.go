package main

import "log"

type Room struct {
	hub     *Hub
	id      string
	usersId []string
	send    chan []byte
}

func (r *Room) writePump() {
	for {
		select {
		case message, ok := <-r.send:
			if !ok {
				log.Println("Room was deleted")
			}
			for _, id := range r.usersId {
				user, ok := r.hub.rooms[id]
				if !ok {
					log.Printf("User: %s, not connected\n", id)
				}
				*user.GetSendChan() <- message
			}
		}
	}
}

func (r *Room) GetSendChan() *chan []byte {
	return &r.send
}

func (r *Room) GetId() string {
	return r.id
}
