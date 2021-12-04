package main

type HubWriter interface {
	GetSendChan() *chan Message
	GetId() string
}
