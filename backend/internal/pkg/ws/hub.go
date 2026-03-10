package ws

import (
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Message represents a WebSocket message sent to clients.
type Message struct {
	Type    string      `json:"type"`
	Data interface{} `json:"data"`
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	// clients maps user IDs to their active WebSocket clients (a user can have multiple connections).
	clients map[uuid.UUID][]*Client

	// register channel for new clients.
	register chan *Client

	// unregister channel for disconnecting clients.
	unregister chan *Client

	// broadcast channel for messages to all clients.
	broadcast chan []byte

	mu sync.RWMutex
}

// NewHub creates a new Hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID][]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
	}
}

// Run starts the hub's main event loop. Should be run as a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = append(h.clients[client.UserID], client)
			h.mu.Unlock()
			log.Debug().
				Str("user_id", client.UserID.String()).
				Msg("ws client registered")

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.UserID]; ok {
				for i, c := range clients {
					if c == client {
						h.clients[client.UserID] = append(clients[:i], clients[i+1:]...)
						break
					}
				}
				if len(h.clients[client.UserID]) == 0 {
					delete(h.clients, client.UserID)
				}
			}
			close(client.send)
			h.mu.Unlock()
			log.Debug().
				Str("user_id", client.UserID.String()).
				Msg("ws client unregistered")

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, clients := range h.clients {
				for _, client := range clients {
					select {
					case client.send <- message:
					default:
						// Client send buffer full; skip.
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// SendToUser sends a message to all active connections for a specific user.
func (h *Hub) SendToUser(userID uuid.UUID, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Warn().Err(err).Msg("failed to marshal ws message")
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.clients[userID]
	if !ok {
		return
	}

	for _, client := range clients {
		select {
		case client.send <- data:
		default:
			// Client send buffer full; skip.
		}
	}
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Warn().Err(err).Msg("failed to marshal ws broadcast message")
		return
	}

	h.broadcast <- data
}

// Register returns the register channel for adding new clients.
func (h *Hub) Register() chan<- *Client {
	return h.register
}

// Unregister returns the unregister channel for removing clients.
func (h *Hub) Unregister() chan<- *Client {
	return h.unregister
}
