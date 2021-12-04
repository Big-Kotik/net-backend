package main

// HubWriter interface for writing message by hub
type HubWriter interface {
	GetSendChan() chan Message
	GetID() string
}
