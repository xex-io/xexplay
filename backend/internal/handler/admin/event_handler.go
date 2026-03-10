package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

type EventHandler struct {
	eventRepo *postgres.EventRepo
}

func NewEventHandler(eventRepo *postgres.EventRepo) *EventHandler {
	return &EventHandler{eventRepo: eventRepo}
}

// List handles GET /admin/events
func (h *EventHandler) List(c *gin.Context) {
	events, err := h.eventRepo.FindAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to fetch events")
		return
	}

	response.OK(c, events)
}

// Create handles POST /admin/events
func (h *EventHandler) Create(c *gin.Context) {
	// TODO: implement full create with validation
	response.OK(c, gin.H{"message": "event creation placeholder"})
}

// Update handles PUT /admin/events/:id
func (h *EventHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	_, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid event id")
		return
	}

	// TODO: implement full update with validation
	response.OK(c, gin.H{"message": "event update placeholder"})
}
