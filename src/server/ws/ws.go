package ws

import (
	"github.com/gorilla/websocket"
	"log"
)


type Client struct {
	socketHub  *SocketHub
	connection *websocket.Conn
	send       chan []byte
}

func NewClient(socketHub *SocketHub, connection *websocket.Conn) *Client {
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

func (s *SocketHub) Run() {
	for message := range s.broadcast {
		for client := range s.clients {
			client.send <- message
		}
	}
}

func NewHub() *SocketHub {
	return &SocketHub{
		clients:   make(map[*Client]bool),
		broadcast: make(chan []byte),
	}
}


func (c *Client) Write() {
	for message := range c.send {
		c.connection.WriteMessage(websocket.TextMessage, message)
	}
}

func (c *Client) Read() {
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