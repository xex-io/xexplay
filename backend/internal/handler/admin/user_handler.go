package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage

	users, err := h.userRepo.ListPaginated(c.Request.Context(), perPage, offset)
	if err != nil {
		response.InternalError(c, "failed to fetch users")
		return
	}

	total, err := h.userRepo.Count(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to count users")
		return
	}

	response.Paginated(c, users, page, perPage, total)
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

type updateUserRequest struct {
	Role     string `json:"role"`
	IsActive *bool  `json:"is_active"`
}

// Update handles PUT /admin/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	// Fetch existing user
	user, err := h.userRepo.FindByID(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "failed to fetch user")
		return
	}
	if user == nil {
		response.NotFound(c, "user not found")
		return
	}

	// Apply updates
	role := user.Role
	if req.Role != "" {
		switch req.Role {
		case domain.RoleUser, domain.RoleAdmin:
			role = req.Role
		default:
			response.BadRequest(c, "invalid role: must be user or admin")
			return
		}
	}

	isActive := user.IsActive
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	if err := h.userRepo.UpdateAdmin(c.Request.Context(), id, role, isActive); err != nil {
		response.InternalError(c, "failed to update user: "+err.Error())
		return
	}

	// Return updated user
	user.Role = role
	user.IsActive = isActive
	response.OK(c, user)
}
