package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

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
	hub := GetHub()
	id := r.URL.Query().Get("id")
	if hub.ContainsID(id) {
		_, err := w.Write([]byte("Ok"))
		if err != nil {
			http.Error(w, "Method crash", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "No such user", http.StatusNotFound)
	}
}

func serveClientWs(hub Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	id := getID()

	client := &Client{hub: hub, conn: conn, send: make(chan ClientMessage, 256), id: id}
	client.hub.Register(client)

	client.send <- ClientMessage{Destination: id, Source: hub.GetID(), Message: "Success"}

	go client.writePump()
	go client.readPump()
}

func serveRoom(hub Hub, w http.ResponseWriter, r *http.Request) {
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

	id := getID()
	room := &Room{hub: hub, id: id, usersID: ids, send: make(chan ClientMessage)}

	hub.Register(room)

	go room.writePump()

	_, err = w.Write([]byte(id))
	if err != nil {
		log.Printf("err: %v", err)
		return
	}
}
