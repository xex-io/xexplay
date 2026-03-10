package handler

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

// GetProfile handles GET /me
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	user, err := h.userRepo.FindByID(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, "failed to fetch profile")
		return
	}
	if user == nil {
		response.NotFound(c, "user not found")
		return
	}

	response.OK(c, user)
}

type updateProfileRequest struct {
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Language    string `json:"language"`
}

// UpdateProfile handles PUT /me
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.userRepo.UpdateProfile(c.Request.Context(), uid, req.DisplayName, req.AvatarURL, req.Language); err != nil {
		response.InternalError(c, "failed to update profile")
		return
	}

	// Fetch updated user
	user, err := h.userRepo.FindByID(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, "failed to fetch updated profile")
		return
	}

	response.OK(c, user)
}

// GetStats handles GET /me/stats
func (h *UserHandler) GetStats(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	stats, err := h.userRepo.GetStats(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, "failed to fetch stats")
		return
	}

	response.OK(c, stats)
}

// GetHistory handles GET /me/history
func (h *UserHandler) GetHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	// TODO: implement history query with pagination
	_ = uid
	response.OK(c, []interface{}{})
}
