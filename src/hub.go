package main

type Message struct {
	Id int `json:"id"`
	Message []byte `json:"message"`
}

type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	//get client from id
	registered map[int]*Client

	// Inbound messages from the clients.
	broadcast chan Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		registered: make(map[int]*Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			//for client := range h.clients {
			//	select {
			//	case client.send <- message:
			//	default:
			//		close(client.send)
			//		delete(h.clients, client)
			//	}
			//}
			if client, err := h.registered[message.Id]; !err {
				select {
				case client.send <- message.Message:
				default:
					close(client.send)
					delete(h.clients, client)
					h.registered[message.Id] = nil
				}
			}
		}
	}
}
