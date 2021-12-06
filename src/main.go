package main

import (
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var port = flag.String("port", ":8080", "http service port")

// type for serve hub functions
type ServeHandler func(Hub, http.ResponseWriter, *http.Request)

func applyServeFunc(hub Hub, serve ServeHandler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) { serve(hub, writer, request) }
}

func newServer() *http.Server {
	r := mux.NewRouter()
	debug := flag.Bool("debug", false, "run debug mode")
	hub := GetHub()
	flag.Parse()
	if *debug {
		r.HandleFunc("/", serveHome)
	}
	r.HandleFunc("/create_room", applyServeFunc(hub, serveRoom))
	r.Path("/check_id").Methods("GET").Queries("id", "{id}").HandlerFunc(serveCheckIDExist)
	r.HandleFunc("/ws/client", applyServeFunc(hub, serveClientWs))

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
