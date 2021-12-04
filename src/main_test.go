package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
	"io"
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
}

func (s *APISuite) SetupSuite() {
	srv := newServer()

	time.Sleep(1000 * time.Millisecond)

	go func() {
		log.Fatal(srv.ListenAndServe())
	}()
}

func (s *APISuite) TestWebSockets() {
	parseID := func(message []byte) string {
		idSlice := make([]Message, 0)
		err := json.Unmarshal(message, &idSlice)
		if err != nil {
			s.Require().Failf("Can't parse message", "fail with error: %v", err)
		}
		s.Require().Equal("Success", idSlice[0].Message)
		return idSlice[0].Destination
	}

	s.Run("two sockets test", func() {
		u := url.URL{Scheme: "ws", Host: servAddr, Path: wsEndpoint}

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

		firstID := parseID(firstMessage)
		secondID := parseID(secondMessage)

		s.Require().NotEqual(firstID, secondID)

		testMessage := Message{firstID, secondID, "Hello, world!"}
		secondSocket.WriteJSON(testMessage)

		_, getMessage, err := firstSocket.ReadMessage()

		if err != nil {
			s.Fail("Error", "err: %v", err)
		}

		messages := make([]Message, 0)
		json.Unmarshal(getMessage, &messages)

		s.Require().Equal(testMessage, messages[0])
	})
}

func (s *APISuite) TestRooms() {
	parseID := func(message []byte) string {
		idSlice := make([]Message, 0)
		err := json.Unmarshal(message, &idSlice)
		if err != nil {
			s.Require().Failf("Can't parse message", "fail with error: %v", err)
		}
		s.Require().Equal("Success", idSlice[0].Message)
		return idSlice[0].Destination
	}
	s.Run("Test writers", func() {
		u := url.URL{Scheme: "ws", Host: servAddr, Path: wsEndpoint}
		sockets := make([]*websocket.Conn, 5)
		ids := make([]string, 5)

		for i := range sockets {
			var err error
			sockets[i], _, err = websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				s.Require().Failf("Can't create socket", "Socket %d, fail with error: %v", i, err)
			}
			sockets[i].SetReadDeadline(time.Now().Add(pongWait))

			_, message, err := sockets[i].ReadMessage()
			if err != nil {
				s.Require().Failf("Can't read message", "fail with error: %v", err)
			}
			ids[i] = parseID(message)
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
		data, err = io.ReadAll(response.Body)
		if err != nil {
			s.Require().Failf("Fail to read response", "fail with error: %v", err)
			return
		}
		roomID := string(data)

		testMessage := Message{roomID, ids[0], "Hello, world!"}
		sockets[0].WriteJSON(testMessage)

		for ind, sock := range sockets {
			_, getMessage, _ := sock.ReadMessage()

			testMessage.Destination = ids[ind]

			messages := make([]Message, 0)
			json.Unmarshal(getMessage, &messages)

			s.Require().Equal(testMessage, messages[0])
		}

	})
}
