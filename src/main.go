package main

import (
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var port = flag.String("port", ":8080", "http service port")

func newServer() *http.Server {
	r := mux.NewRouter()
	debug := flag.Bool("debug", false, "run debug mode")
	hub := GetHub()
	flag.Parse()
	if *debug {
		r.HandleFunc("/", serveHome)
	}
	r.HandleFunc("/create_room", func(writer http.ResponseWriter, request *http.Request) {
		serveRoom(hub, writer, request)
	})
	r.HandleFunc("/ws/client", func(w http.ResponseWriter, r *http.Request) {
		serveClientWs(hub, w, r)
	})

	return &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0" + *port,
		WriteTimeout: writeWait,
		ReadTimeout:  writeWait,
	}
}

func main() {
	srv := newServer()
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("can't create server")
		return
	}
}
