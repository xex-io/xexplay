package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

type DeviceHandler struct {
	fcmRepo *postgres.FCMTokenRepo
}

func NewDeviceHandler(fcmRepo *postgres.FCMTokenRepo) *DeviceHandler {
	return &DeviceHandler{fcmRepo: fcmRepo}
}

type registerDeviceRequest struct {
	Token      string `json:"token" binding:"required"`
	DeviceType string `json:"device_type" binding:"required,oneof=ios android web"`
}

// Register handles POST /devices/register — registers an FCM token for push notifications.
func (h *DeviceHandler) Register(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	var req registerDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: token and device_type (ios/android/web) required")
		return
	}

	token := &domain.FCMToken{
		ID:         uuid.New(),
		UserID:     uid,
		Token:      req.Token,
		DeviceType: req.DeviceType,
	}

	if err := h.fcmRepo.RegisterToken(c.Request.Context(), token); err != nil {
		response.InternalError(c, "failed to register device token")
		return
	}

	response.Created(c, token)
}

// Deregister handles DELETE /devices/:token — deactivates an FCM token.
func (h *DeviceHandler) Deregister(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	tokenParam := c.Param("token")
	if tokenParam == "" {
		response.BadRequest(c, "token is required")
		return
	}

	if err := h.fcmRepo.DeactivateToken(c.Request.Context(), uid, tokenParam); err != nil {
		response.NotFound(c, "token not found")
		return
	}

	response.OK(c, gin.H{"message": "device token deactivated"})
}
