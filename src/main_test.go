package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

var (
	servAddr   = "0.0.0.0:8080"
	wsEndpoint = "/ws"
)

func Test(t *testing.T) {
	suite.Run(t, &APISuite{})
}

type APISuite struct {
	suite.Suite

	cl http.Client
}

func (s *APISuite) SetupSuite() {
	srv := newServer()

	time.Sleep(1000 * time.Millisecond)

	go func() {
		log.Fatal(srv.ListenAndServe())
	}()
}

func (s *APISuite) TestWebSockets() {
	parseId := func(message []byte) string {
		idSlice := make([]string, 0)
		err := json.Unmarshal(message, &idSlice)
		if err != nil {
			s.Require().Failf("Can't parse message", "fail with error: %v", err)
		}
		return idSlice[0]
	}

	s.Run("two sockets test", func() {
		u := url.URL{Scheme: "ws", Host: servAddr, Path: "/ws"}

		firstSocket, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			s.Require().Failf("Can't create socket", "fail with error: %v", err)
		}

		secondSocket, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			s.Require().Failf("Can't create socket", "fail with error: %v", err)
		}

		defer firstSocket.Close()
		defer secondSocket.Close()
		firstSocket.SetReadDeadline(time.Now().Add(pongWait))
		secondSocket.SetReadDeadline(time.Now().Add(pongWait))

		_, firstMessage, err := firstSocket.ReadMessage()
		if err != nil {
			s.Require().Failf("Can't read message", "fail with error: %v", err)
		}

		_, secondMessage, err := secondSocket.ReadMessage()
		if err != nil {
			s.Require().Failf("Can't read message", "fail with error: %v", err)
		}

		firstId := parseId(firstMessage)
		secondId := parseId(secondMessage)

		s.Require().NotEqual(firstId, secondId)

		secondSocket.WriteJSON(Message{firstId, "Hello, world!"})

		_, getMessage, err := firstSocket.ReadMessage()

		s.Require().Equal("[\"Hello, world!\"]", string(getMessage))
	})
}

func (s *APISuite) TestRooms() {
	parseId := func(message []byte) string {
		idSlice := make([]string, 0)
		err := json.Unmarshal(message, &idSlice)
		if err != nil {
			s.Require().Failf("Can't parse message", "fail with error: %v", err)
		}
		return idSlice[0]
	}
	s.Run("Test rooms", func() {
		u := url.URL{Scheme: "ws", Host: servAddr, Path: "/ws"}
		sockets := make([]*websocket.Conn, 5)
		ids := make([]string, 5)

		for i := range sockets {
			var err error
			sockets[i], _, err = websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				s.Require().Failf("Can't create socket", "fail with error: %v", err)
			}
			sockets[i].SetReadDeadline(time.Now().Add(pongWait))

			_, message, err := sockets[i].ReadMessage()
			if err != nil {
				s.Require().Failf("Can't read message", "fail with error: %v", err)
			}
			ids[i] = parseId(message)
		}
		defer func() {
			for _, sock := range sockets {
				sock.Close()
			}
		}()

		makeRoom := url.URL{Scheme: "http", Host: servAddr, Path: "/create_room"}
		data, _ := json.Marshal(ids)
		request, err := http.NewRequest("POST", makeRoom.String(), strings.NewReader(string(data)))
		if err != nil {
			s.Require().Failf("Can't create request", "fail with error: %v", err)
			return
		}

		client := &http.Client{}
		response, err := client.Do(request)
		if err != nil {
			s.Require().Failf("Fail to create room", "fail with error: %v", err)
			return
		}
		data, err = ioutil.ReadAll(response.Body)
		if err != nil {
			s.Require().Failf("Fail to read response", "fail with error: %v", err)
			return
		}
		roomId := string(data)

		sockets[0].WriteJSON(Message{roomId, "Hello, world!"})

		for _, sock := range sockets {
			_, getMessage, _ := sock.ReadMessage()
			s.Require().Equal("[\"Hello, world!\"]", string(getMessage))
		}

	})
}
