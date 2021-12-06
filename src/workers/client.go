package workers

import (
	"encoding/json"
	"log"
	hub2 "net-backend/src/hub"
	"net-backend/src/msg"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// WriteWait Time allowed to write a msg to the peer.
	WriteWait = 10 * time.Second

	// PongWait Time allowed to read the next pong msg from the peer.
	PongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than PongWait.
	pingPeriod = (PongWait * 9) / 10

	// Maximum msg size allowed from peer.
	maxMessageSize = 512
)

// Client struct for client
type Client struct {
	Hub hub2.Hub

	Conn *websocket.Conn

	Send chan msg.ClientMessage

	ID string
}

// GetSendChan implementation of Client.GetSendChan()
func (c *Client) GetSendChan() chan msg.ClientMessage {
	return c.Send
}

// GetID implementation of Client.GetID()
func (c *Client) GetID() string {
	return c.ID
}

// ReadPump read websocket
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		err := c.Conn.Close()
		if err != nil {
			return
		}
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	err := c.Conn.SetReadDeadline(time.Now().Add(PongWait))
	if err != nil {
		log.Println("Can't set read deadline")
		return
	}
	c.Conn.SetPongHandler(func(string) error {
		err := c.Conn.SetReadDeadline(time.Now().Add(PongWait))
		return err
	})
	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		var message msg.ClientMessage
		err = json.Unmarshal(messageBytes, &message)

		if err != nil {
			log.Printf("error: %v", err)
			log.Printf("can't parse to json")
			break
		}

		c.Hub.SendMessage(message)
	}
}

// WritePump write to websocket
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := c.Conn.Close()
		if err != nil {
			log.Println("Error while close connection")
			return
		}
	}()
	for {
		select {
		case clientMessage, ok := <-c.Send:
			err := c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err != nil {
				log.Printf("err: %v", err)
				return
			}
			if !ok {
				err = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Printf("err: %v", err)
					return
				}
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Println("err")
				return
			}
			n := len(c.Send)
			messages := make([]msg.ClientMessage, n+1)

			messages[0] = clientMessage
			for i := 1; i < n+1; i++ {
				messages[i] = <-c.Send
			}

			messageJSON, _ := json.Marshal(messages)
			_, err = w.Write(messageJSON)
			if err != nil {
				log.Printf("err: %v", err)
				return
			}

			if err := w.Close(); err != nil {
				log.Println("err")
				return
			}

		case <-ticker.C:
			err := c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err != nil {
				log.Printf("err: %v", err)
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
