package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	jwtpkg "github.com/xex-exchange/xexplay-api/internal/pkg/jwt"
	wspkg "github.com/xex-exchange/xexplay-api/internal/pkg/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, restrict this to allowed origins.
		return true
	},
}

// WebSocketHandler handles WebSocket upgrade requests.
type WebSocketHandler struct {
	hub       *wspkg.Hub
	jwtSecret string
}

// NewWebSocketHandler creates a new WebSocketHandler.
func NewWebSocketHandler(hub *wspkg.Hub, jwtSecret string) *WebSocketHandler {
	return &WebSocketHandler{
		hub:       hub,
		jwtSecret: jwtSecret,
	}
}

// Handle upgrades the HTTP connection to a WebSocket connection.
// Authentication is done via the "token" query parameter.
// GET /ws?token=<jwt_token>
func (h *WebSocketHandler) Handle(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token query parameter"})
		return
	}

	claims, err := jwtpkg.Parse(token, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Warn().Err(err).Msg("failed to upgrade websocket connection")
		return
	}

	client := wspkg.NewClient(h.hub, conn, claims.UserID)

	h.hub.Register() <- client

	go client.WritePump()
	go client.ReadPump()
}
