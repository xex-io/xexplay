package handler

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

// List handles GET /events — returns active events for public users.
func (h *EventHandler) List(c *gin.Context) {
	events, err := h.eventRepo.FindActive(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to fetch events")
		return
	}

	response.OK(c, events)
}

// GetByID handles GET /events/:id
func (h *EventHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid event id")
		return
	}

	event, err := h.eventRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "failed to fetch event")
		return
	}
	if event == nil {
		response.NotFound(c, "event not found")
		return
	}

	response.OK(c, event)
}
