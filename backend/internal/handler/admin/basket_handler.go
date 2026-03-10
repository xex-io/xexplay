package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
)

type BasketHandler struct{}

func NewBasketHandler() *BasketHandler {
	return &BasketHandler{}
}

// List handles GET /admin/baskets
func (h *BasketHandler) List(c *gin.Context) {
	// TODO: implement with pagination and optional date filter
	response.OK(c, []interface{}{})
}

// Create handles POST /admin/baskets
func (h *BasketHandler) Create(c *gin.Context) {
	// TODO: implement full create with card assignment
	response.OK(c, gin.H{"message": "basket creation placeholder"})
}

// Update handles PUT /admin/baskets/:id
func (h *BasketHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	_, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid basket id")
		return
	}

	// TODO: implement full update with card reassignment
	response.OK(c, gin.H{"message": "basket update placeholder"})
}

// Publish handles POST /admin/baskets/:id/publish
func (h *BasketHandler) Publish(c *gin.Context) {
	idParam := c.Param("id")
	_, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid basket id")
		return
	}

	// TODO: implement publish (set is_published = true, validate 15 cards assigned)
	response.OK(c, gin.H{"message": "basket publish placeholder"})
}
