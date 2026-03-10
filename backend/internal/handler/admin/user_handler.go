package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

type UserHandler struct {
	userRepo *postgres.UserRepo
}

func NewUserHandler(userRepo *postgres.UserRepo) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

// List handles GET /admin/users
func (h *UserHandler) List(c *gin.Context) {
	// TODO: implement with pagination, search, and filtering
	response.OK(c, []interface{}{})
}

// GetByID handles GET /admin/users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	user, err := h.userRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "failed to fetch user")
		return
	}
	if user == nil {
		response.NotFound(c, "user not found")
		return
	}

	response.OK(c, user)
}

// Update handles PUT /admin/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	_, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	// TODO: implement admin user update (role, is_active, etc.)
	response.OK(c, gin.H{"message": "user update placeholder"})
}
