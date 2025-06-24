package main

import (
	"log"
	"net/http"
	"server/storage"
	"server/ws"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func serveWs(socketHub *ws.SocketHub, w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := ws.NewClient(socketHub, conn)
	go client.Read()
	go client.Write()
}

func main() {
	socketHub := ws.NewHub()
	clientStorage := storage.NewDeviceStorage()
	clientStorage.AddDevice()
	clientStorage.GetAllDevices()

	go socketHub.Run()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(socketHub, w, r)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
