package notification

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all cross-origin for local & cloud POS clients
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	Hub    *WebSocketHub
	Conn   *websocket.Conn
	Send   chan []byte
	OrgID  int32
	UserID int32
}

type WebSocketHub struct {
	clients    map[*Client]bool
	orgClients map[int32]map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[*Client]bool),
		orgClients: make(map[int32]map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *WebSocketHub) Run(stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			h.mu.Lock()
			for client := range h.clients {
				close(client.Send)
				_ = client.Conn.Close()
			}
			h.clients = make(map[*Client]bool)
			h.orgClients = make(map[int32]map[*Client]bool)
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			if _, ok := h.orgClients[client.OrgID]; !ok {
				h.orgClients[client.OrgID] = make(map[*Client]bool)
			}
			h.orgClients[client.OrgID][client] = true
			h.mu.Unlock()
			log.Printf("[WebSocketHub] Client registered: OrgID=%d UserID=%d", client.OrgID, client.UserID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				if orgMap, exists := h.orgClients[client.OrgID]; exists {
					delete(orgMap, client)
					if len(orgMap) == 0 {
						delete(h.orgClients, client.OrgID)
					}
				}
				close(client.Send)
				_ = client.Conn.Close()
			}
			h.mu.Unlock()
			log.Printf("[WebSocketHub] Client disconnected: OrgID=%d UserID=%d", client.OrgID, client.UserID)

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToOrg sends a JSON payload to all connected clients in a specific organization
func (h *WebSocketHub) BroadcastToOrg(orgID int32, payload interface{}) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	orgMap, ok := h.orgClients[orgID]
	if !ok {
		return
	}

	for client := range orgMap {
		select {
		case client.Send <- bytes:
		default:
			// Buffer full, skip
		}
	}
}

// SendToUser sends a JSON payload to a specific user in an organization
func (h *WebSocketHub) SendToUser(orgID, userID int32, payload interface{}) {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	orgMap, ok := h.orgClients[orgID]
	if !ok {
		return
	}

	for client := range orgMap {
		if client.UserID == userID {
			select {
			case client.Send <- bytes:
			default:
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
	}()
	c.Conn.SetReadLimit(512)
	_ = c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles WebSocket handshake requests
func (h *WebSocketHub) ServeWs(c *gin.Context, orgID, userID int32) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WebSocketHub] Upgrade failed: %v", err)
		return
	}

	client := &Client{
		Hub:    h,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		OrgID:  orgID,
		UserID: userID,
	}

	h.register <- client

	go client.writePump()
	go client.readPump()
}
