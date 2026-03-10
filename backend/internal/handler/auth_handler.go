package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type loginRequest struct {
	Token        string `json:"token" binding:"required"`
	ReferralCode string `json:"referral_code,omitempty"`
	DeviceID     string `json:"device_id,omitempty"`
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "token is required")
		return
	}

	ip := c.ClientIP()
	user, err := h.authService.Login(c.Request.Context(), req.Token, req.ReferralCode, req.DeviceID, ip)
	if err != nil {
		response.Unauthorized(c, "authentication failed")
		return
	}

	response.OK(c, user)
}
