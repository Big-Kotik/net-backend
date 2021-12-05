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
	servAddr         = "0.0.0.0:8080"
	wsClientEndpoint = "/ws/client"
	checkIDEndpoint  = "/check_id"
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
		idSlice := make([]ClientMessage, 0)
		err := json.Unmarshal(message, &idSlice)
		if err != nil {
			s.Require().Failf("Can't parse message", "fail with error: %v", err)
		}
		s.Require().Equal("Success", idSlice[0].Message)
		return idSlice[0].Destination
	}

	s.Run("two sockets test", func() {
		u := url.URL{Scheme: "ws", Host: servAddr, Path: wsClientEndpoint}

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

		testMessage := ClientMessage{firstID, secondID, "Hello, world!"}
		secondSocket.WriteJSON(testMessage)

		_, getMessage, err := firstSocket.ReadMessage()

		if err != nil {
			s.Fail("Error", "err: %v", err)
		}

		messages := make([]ClientMessage, 0)
		json.Unmarshal(getMessage, &messages)

		s.Require().Equal(testMessage, messages[0])
	})
}

func (s *APISuite) TestRooms() {
	parseID := func(message []byte) string {
		idSlice := make([]ClientMessage, 0)
		err := json.Unmarshal(message, &idSlice)
		s.NoErrorf(err, "Fail err %v", err)
		s.Require().Equal("Success", idSlice[0].Message)
		return idSlice[0].Destination
	}
	s.Run("Test writers", func() {
		u := url.URL{Scheme: "ws", Host: servAddr, Path: wsClientEndpoint}
		sockets := make([]*websocket.Conn, 5)
		ids := make([]string, 5)

		for i := range sockets {
			var err error
			sockets[i], _, err = websocket.DefaultDialer.Dial(u.String(), nil)
			s.NoErrorf(err, "Socket %d, fail with error: %v", i, err)
			sockets[i].SetReadDeadline(time.Now().Add(pongWait))
			_, message, err := sockets[i].ReadMessage()
			s.NoErrorf(err, "fail with error: %v", err)
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

		testMessage := ClientMessage{roomID, ids[0], "Hello, world!"}
		sockets[0].WriteJSON(testMessage)

		for ind, sock := range sockets {
			_, getMessage, _ := sock.ReadMessage()

			testMessage.Destination = ids[ind]

			messages := make([]ClientMessage, 0)
			json.Unmarshal(getMessage, &messages)

			s.Require().Equal(testMessage, messages[0])
		}

	})
}

func (s *APISuite) TestUserExist() {
	parseID := func(message []byte) string {
		idSlice := make([]ClientMessage, 0)
		err := json.Unmarshal(message, &idSlice)
		s.NoErrorf(err, "Fail err %v", err)
		s.Require().Equal("Success", idSlice[0].Message)
		return idSlice[0].Destination
	}
	s.Run("Test /check_id", func() {
		wsURL := url.URL{Scheme: "ws", Host: servAddr, Path: wsClientEndpoint}
		socket, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
		s.NoError(err)
		defer socket.Close()

		socket.SetReadDeadline(time.Now().Add(pongWait))
		_, message, err := socket.ReadMessage()
		s.NoError(err)
		id := parseID(message)

		checkURL := url.URL{
			Scheme:   "http",
			Host:     servAddr,
			Path:     checkIDEndpoint,
			RawQuery: "id=" + id,
		}

		request, err := http.NewRequest("GET", checkURL.String(), strings.NewReader(""))
		s.NoError(err)
		client := &http.Client{}
		response, err := client.Do(request)
		s.NoError(err)
		s.Equal(http.StatusOK, response.StatusCode)
		err = socket.Close()
		s.NoError(err)

		request, err = http.NewRequest("GET", checkURL.String(), strings.NewReader(""))
		s.NoError(err)
		response, err = client.Do(request)
		s.NoError(err)
		s.Equal(http.StatusNotFound, response.StatusCode)
	})
}
