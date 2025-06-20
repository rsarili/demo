package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type Client struct {
	socketHub  *SocketHub
	connection *websocket.Conn
	send       chan []byte
}

func newClient(socketHub *SocketHub, connection *websocket.Conn) *Client {
	client := &Client{
		socketHub:  socketHub,
		connection: connection,
		send:       make(chan []byte),
	}
	socketHub.clients[client] = true
	return client
}

type SocketHub struct {
	clients   map[*Client]bool
	broadcast chan []byte
}

func (s *SocketHub) run() {
	for message := range s.broadcast {
		for client := range s.clients {
			client.send <- message
		}
	}
}

func newHub() *SocketHub {
	return &SocketHub{
		clients:   make(map[*Client]bool),
		broadcast: make(chan []byte),
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (c *Client) write() {
	for message := range c.send {
		c.connection.WriteMessage(websocket.TextMessage, message)
	}
}

func (c *Client) read() {
	defer func() {
		c.connection.Close()
	}()

	c.connection.WriteMessage(websocket.TextMessage, []byte("Hello world"))

	for {
		_, message, err := c.connection.ReadMessage()
		log.Printf("incoming message: %s", message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		c.socketHub.broadcast <- message
	}
}

func serveWs(socketHub *SocketHub, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := newClient(socketHub, conn)
	go client.read()
	go client.write()
}

func main() {
	socketHub := newHub()
	go socketHub.run()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(socketHub, w, r)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
