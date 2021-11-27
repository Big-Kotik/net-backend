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
	//debug := flag.Bool("debug", false, "run debug mode")
	hub := newHub()
	flag.Parse()

	go hub.run()
	if true {
		r.HandleFunc("/", serveHome)
	}
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	//err := http.ListenAndServe(*addr, nil)
	//if err != nil {
	//	log.Fatal("ListenAndServe: ", err)
	//}

	return &http.Server{
		Handler: r,
		Addr: "0.0.0.0:8080",
		WriteTimeout: writeWait,
		ReadTimeout: writeWait,
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
