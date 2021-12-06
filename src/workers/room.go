package workers

import (
	"log"
	hub "net-backend/src/hub"
	"net-backend/src/msg"
)

// Room struct for rooms support
type Room struct {
	Hub     hub.Hub
	ID      string
	UsersID []string
	Send    chan msg.ClientMessage
}

// WritePump write message which came to Send channel
func (r *Room) WritePump() {
	for {
		message, ok := <-r.Send
		if !ok {
			log.Println("Room was deleted")
		}
		for _, id := range r.UsersID {
			message.Destination = id
			r.Hub.SendMessage(message)
		}
	}
}

// GetSendChan implementation of Client.GetSendChan()
func (r *Room) GetSendChan() chan msg.ClientMessage {
	return r.Send
}

// GetID implementation of Client.GetID()
func (r *Room) GetID() string {
	return r.ID
}
