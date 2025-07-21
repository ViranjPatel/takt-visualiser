package api

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type WebSocketHandler struct {
	redis    *redis.Client
	upgrader websocket.Upgrader
}

func NewWebSocketHandler(redis *redis.Client) *WebSocketHandler {
	return &WebSocketHandler{
		redis: redis,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in dev
			},
		},
	}
}

func (h *WebSocketHandler) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Subscribe to Redis channel
	pubsub := h.redis.Subscribe(r.Context(), "task_updates")
	defer pubsub.Close()

	// Channel for pings
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case msg := <-pubsub.Channel():
			err := conn.WriteJSON(map[string]interface{}{
				"type":    "task_update",
				"task_id": msg.Payload,
			})
			if err != nil {
				return
			}
		case <-pingTicker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
