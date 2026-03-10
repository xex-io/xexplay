package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

type MatchHandler struct {
	matchRepo *postgres.MatchRepo
}

func NewMatchHandler(matchRepo *postgres.MatchRepo) *MatchHandler {
	return &MatchHandler{matchRepo: matchRepo}
}

// List handles GET /admin/matches
func (h *MatchHandler) List(c *gin.Context) {
	// TODO: implement with pagination and optional event_id filter
	response.OK(c, []interface{}{})
}

// Create handles POST /admin/matches
func (h *MatchHandler) Create(c *gin.Context) {
	// TODO: implement full create with validation
	response.OK(c, gin.H{"message": "match creation placeholder"})
}

// Update handles PUT /admin/matches/:id
func (h *MatchHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	_, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid match id")
		return
	}

	// TODO: implement full update with validation
	response.OK(c, gin.H{"message": "match update placeholder"})
}
