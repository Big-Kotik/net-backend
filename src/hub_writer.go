package main

type HubWriter interface {
	GetSendChan() *chan []byte
	GetId() string
}
