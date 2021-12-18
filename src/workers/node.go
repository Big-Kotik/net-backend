package workers

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net-backend/src/hub"
	"net-backend/src/msg"
	"time"
)

// Node implement Node type
type Node struct {
	Hub hub.Hub

	Conn *websocket.Conn

	Send chan msg.NodeMessage

	Worker chan msg.NodeMessage

	ID string
}

// GetSendChan return channel for sending message
func (n *Node) GetSendChan() chan msg.NodeMessage {
	return n.Send
}

// GetID return ID
func (n *Node) GetID() string {
	return n.ID
}

// Register node
func (n *Node) Register() {
	message := msg.NodeMessage{
		IsAnswer: false,
		Type:     msg.GetNodeID,
		Source:   n.Hub.GetID(),
	}
	n.Send <- message

	for n.ID == "" {
		time.Sleep(time.Second)
	}

	n.Hub.RegisterNode(n)
}

// ReadPump read websocket
func (n *Node) ReadPump() {
	defer func() {
		if err := n.Conn.Close(); err != nil {
			return
		}
	}()
	n.Conn.SetReadLimit(maxMessageSize)
	if err := n.Conn.SetReadDeadline(time.Now().Add(PongWait)); err != nil {
		log.Println("Can't set read deadline")
		return
	}

	n.Conn.SetPongHandler(func(string) error {
		err := n.Conn.SetReadDeadline(time.Now().Add(PongWait))
		return err
	})

	for {
		_, messageBytes, err := n.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		var message msg.NodeMessage
		err = json.Unmarshal(messageBytes, &message)

		if err != nil {
			log.Printf("error: %v", err)
			log.Printf("can't parse to json")
			break
		}

		n.Worker <- message
	}
}

// WritePump write messages to websocket
func (n *Node) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := n.Conn.Close(); err != nil {
			log.Println("Error while close connection")
			return
		}
	}()
	for {
		select {
		case message, ok := <-n.Send:
			if err := n.Conn.SetWriteDeadline(time.Now().Add(WriteWait)); err != nil {
				log.Printf("err: %v", err)
				return
			}
			if !ok {
				if err := n.Conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.Printf("err: %v", err)
					return
				}
				return
			}
			w, err := n.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Println("error")
				return
			}

			count := len(n.Send)
			messages := make([]msg.NodeMessage, count+1)

			messages[0] = message
			for i := 1; i < count+1; i++ {
				messages[i] = <-n.Send
			}

			messageJSON, _ := json.Marshal(messages)
			if _, err = w.Write(messageJSON); err != nil {
				log.Printf("err: %v", err)
				return
			}

			if err := w.Close(); err != nil {
				log.Printf("err: %v", err)
				return
			}
		case <-ticker.C:
			err := n.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err != nil {
				log.Printf("err: %v", err)
				return
			}
			if err := n.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func makeAnswer(requestMessage msg.NodeMessage, bodyValue func() (string, error)) msg.NodeMessage {
	responseMessage := msg.NodeMessage{
		Destination: requestMessage.Source,
		Source:      requestMessage.Destination,
		Type:        requestMessage.Type,
		IsAnswer:    true,
		NodeThrough: 255,
	}
	body, err := bodyValue()
	if err != nil {
		responseMessage.IsError = true
		responseMessage.Body = err.Error()
	} else {
		responseMessage.Body = body
	}
	return responseMessage
}

// Work run worker for message processing
func (n *Node) Work() {
	for {
		message, ok := <-n.Worker
		if !ok {
			log.Println("Worker channel was closed")
			break
		}
		if message.Type == msg.GetNodeID {
			if message.IsAnswer {
				n.ID = message.Body
			} else {
				n.Send <- makeAnswer(message, func() (string, error) {
					m, err := json.Marshal(n.Hub.GetID())
					return string(m), err
				})
			}
		} else if message.Destination != n.Hub.GetID() && !n.Hub.ContainsID(message.Destination) {
			if message.NodeThrough > 1 {
				message.NodeThrough--
				myID := make(map[string]struct{})
				myID[n.ID] = struct{}{}
				n.Hub.BroadcastNodeMessage(message, myID)
			}
		} else if !message.IsAnswer {
			switch message.Type {
			case msg.GetNodesAddress:
				n.Send <- makeAnswer(message, func() (string, error) {
					m, err := json.Marshal(n.Hub.GetNodeIDs())
					return string(m), err
				})
			case msg.GetClientsIDs:
				n.Send <- makeAnswer(message, func() (string, error) {
					m, err := json.Marshal(n.Hub.GetClientIDs())
					return string(m), err
				})
			case msg.ForwardClient:
				var clientMessage msg.ClientMessage
				err := json.Unmarshal([]byte(message.Body), &clientMessage)
				if err != nil {
					continue
				}
				n.Hub.SendMessage(clientMessage)
			}
		}
	}
}
