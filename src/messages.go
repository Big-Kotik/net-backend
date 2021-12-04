package main

// ClientMessage struct for message which came from client
type ClientMessage struct {
	Destination string `json:"destination"`
	Source      string `json:"source"`
	Message     string `json:"message"`
}
