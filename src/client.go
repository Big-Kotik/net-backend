package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client struct for client
type Client struct {
	hub *Hub

	conn *websocket.Conn

	send chan Message

	id string
}

// GetSendChan implementation of HubWriter.GetSendChan()
func (c *Client) GetSendChan() chan Message {
	return c.send
}

// GetID implementation of HubWriter.GetID()
func (c *Client) GetID() string {
	return c.id
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		err := c.conn.Close()
		if err != nil {
			return
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		log.Println("Can't set read deadline")
		return
	}
	c.conn.SetPongHandler(func(string) error {
		err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return err
	})
	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		var message Message
		err = json.Unmarshal(messageBytes, &message)

		if err != nil {
			log.Printf("error: %v", err)
			log.Printf("can't parse to json")
			break
		}

		c.hub.broadcast <- message
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		err := c.conn.Close()
		if err != nil {
			log.Println("Error while close connection")
			return
		}
	}()
	for {
		select {
		case message, ok := <-c.send:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				log.Printf("err: %v", err)
				return
			}
			if !ok {
				err = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					log.Printf("err: %v", err)
					return
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Println("err")
				return
			}
			n := len(c.send)
			messages := make([]Message, n+1)

			messages[0] = message
			for i := 1; i < n+1; i++ {
				messages[i] = <-c.send
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
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				log.Printf("err: %v", err)
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
