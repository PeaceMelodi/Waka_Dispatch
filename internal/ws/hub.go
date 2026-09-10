package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	mu          sync.RWMutex
	connections map[*websocket.Conn]bool
}

func NewHub() *Hub {
	return &Hub{
		connections: make(map[*websocket.Conn]bool),
	}
}

func (h *Hub) Add(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.connections[conn] = true
}

func (h *Hub) Remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.connections, conn)
	conn.Close()
}

func (h *Hub) Broadcast(eventType string, data interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	message := map[string]interface{}{
		"type": eventType,
		"data": data,
	}

	payload, err := json.Marshal(message)
	if err != nil {
		return
	}

	for conn := range h.connections {
		conn.WriteMessage(websocket.TextMessage, payload)
	}
}

func (h *Hub) BroadcastRiderLocation(riderID string, lat float64, lng float64) {
	h.Broadcast("rider_location", map[string]interface{}{
		"rider_id": riderID,
		"lat":      lat,
		"lng":      lng,
	})
}

func (h *Hub) BroadcastOrderStatus(orderID string, status string) {
	h.Broadcast("order_status", map[string]interface{}{
		"order_id": orderID,
		"status":   status,
	})
}