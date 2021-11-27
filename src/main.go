package main

import (
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8080", "http service address")

func newServer() *http.Server {
	r := mux.NewRouter()
	debug := flag.Bool("debug", false, "run debug mode")
	room := flag.Bool("room", false, "run room support")
	hub := newHub()
	flag.Parse()

	go hub.run()
	if *debug {
		r.HandleFunc("/", serveHome)
	}
	if *room {
		r.HandleFunc("/create_room", func(writer http.ResponseWriter, request *http.Request) {
			serveRoom(hub, writer, request)
		})
	}
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	//err := http.ListenAndServe(*addr, nil)
	//if err != nil {
	//	log.Fatal("ListenAndServe: ", err)
	//}

	return &http.Server{
		Handler:      r,
		Addr:         "0.0.0.0:8080",
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
