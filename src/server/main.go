package main

import (
	"encoding/json"
	"log"
	"net/http"
	"server/storage"
	"server/ws"

	"github.com/gorilla/websocket"
)

type Device struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

func main() {
	socketHub := ws.NewHub()
	go socketHub.Run()

	clientStorage := storage.NewDeviceStorage()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(socketHub, w, r)
	})

	http.HandleFunc("/devices", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			devices, err := clientStorage.GetAllDevices()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			response := make([]Device, 0)
			for _, v := range devices {
				response = append(response, Device{Id: v.Id, Type: v.Type})
			}

			response_byte, err := json.Marshal(response)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			w.Write(response_byte)
		default:
			http.Error(w, "not supported method", http.StatusMethodNotAllowed)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

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
