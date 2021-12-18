package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"net-backend/src/hub"
	"net-backend/src/msg"
	"net-backend/src/security"
	"net-backend/src/workers"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "./static/home.html")
}

func serveCheckIDExist(w http.ResponseWriter, r *http.Request) {
	h := hub.GetHub()
	id := r.URL.Query().Get("id")
	if h.ContainsID(id) {
		_, err := w.Write([]byte("Ok"))
		if err != nil {
			http.Error(w, "Method crash", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "No such user", http.StatusNotFound)
	}
}

func serveClientWs(h hub.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	id := security.GetID()

	client := &workers.Client{Hub: h, Conn: conn, Send: make(chan msg.ClientMessage, 256), ID: id}
	client.Hub.RegisterClient(client)

	client.Send <- msg.ClientMessage{Destination: id, Source: h.GetID(), Message: "Success"}

	go client.WritePump()
	go client.ReadPump()
}

func serveNodeWs(h hub.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	node := workers.Node{
		Hub:  h,
		Conn: conn,
		Send: make(chan msg.NodeMessage, 256),
	}

	go node.Work()
	go node.WritePump()
	go node.ReadPump()
	go node.Register()
}

func serveRoom(h hub.Hub, w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	ids := make([]string, 0)
	body, err := io.ReadAll(r.Body)

	if err != nil {
		http.Error(w, "No body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &ids)
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, "Parse error", http.StatusBadRequest)
		return
	}

	id := security.GetID()
	room := &workers.Room{Hub: h, ID: id, UsersID: ids, Send: make(chan msg.ClientMessage)}

	h.RegisterClient(room)

	go room.WritePump()

	_, err = w.Write([]byte(id))
	if err != nil {
		log.Printf("err: %v", err)
		return
	}
}
