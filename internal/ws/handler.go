package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true 
	},
}

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

func (h *Handler) HandleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}

	h.hub.Add(conn)
	log.Println("new client connected to /ws/dispatch")

	go func() {
		defer func() {
			h.hub.Remove(conn)
			log.Println("client disconnected from /ws/dispatch")
		}()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}