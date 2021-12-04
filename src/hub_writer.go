package main

// HubWriter interface for writing message by hub
type HubWriter interface {
	GetSendChan() chan ClientMessage
	GetID() string
}
