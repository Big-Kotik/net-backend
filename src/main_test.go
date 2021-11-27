package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"
)

var (
	servAddr   = "0.0.0.0:8080"
	wsEndpoint = "/ws"
)

func Test(t *testing.T) {
	log.Println("test started")
	suite.Run(t, &APISuite{})
}

type APISuite struct {
	suite.Suite

	cl http.Client
}

func (s *APISuite) SetupSuite() {
	srv := newServer()
	log.Println("setup start")
	go func() {
		log.Println("test server started")
		log.Fatal(srv.ListenAndServe())
	}()
	log.Println("setup end")
}

func (s *APISuite) TestWebSockets() {
	log.Println("two sockets test start")
	s.Run("two sockets test", func() {
		u := url.URL{Scheme: "ws", Host: servAddr, Path: "/ws"}
		firstSocket, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		secondSocket, _, _ := websocket.DefaultDialer.Dial(u.String(), nil)

		if err != nil {
			log.Println(u.String())
			log.Println(err)
			s.Require().Fail("can't create socket")
		}

		defer firstSocket.Close()
		defer secondSocket.Close()
		firstSocket.SetReadDeadline(time.Now().Add(pongWait))
		_, firstMessage, e1 := firstSocket.ReadMessage()
		_, secondMessage, e2 := secondSocket.ReadMessage()

		if e1 != nil || e2 != nil {
			log.Fatal("read message error")
			s.Fail("can't read message")
		}

		var firstId string
		var secondId string
		firstIdSlice := make([]string, 0)
		parseErr := json.Unmarshal(firstMessage, &firstIdSlice)
		log.Println(parseErr)
		s.Require().Equal(parseErr, nil)
		firstId = firstIdSlice[0]
		secondId = string(secondMessage)

		s.Require().NotEqual(firstId, secondId)

		secondSocket.WriteJSON(Message{firstId, "abcd"})

		_, getMessage, err := firstSocket.ReadMessage()

		s.Require().Equal("[\"abcd\"]", string(getMessage))
	})
}
