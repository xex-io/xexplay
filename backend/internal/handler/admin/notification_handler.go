package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/middleware"
	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
	historyRepo         *postgres.NotificationHistoryRepo
}

func NewNotificationHandler(notificationService *service.NotificationService, historyRepo *postgres.NotificationHistoryRepo) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService, historyRepo: historyRepo}
}

type sendNotificationRequest struct {
	Title  string `json:"title" binding:"required"`
	Body   string `json:"body" binding:"required"`
	Target string `json:"target" binding:"required,oneof=all active new dormant"`
}

// Send handles POST /admin/notifications/send — broadcasts a notification.
func (h *NotificationHandler) Send(c *gin.Context) {
	var req sendNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: title, body, and target (all) required")
		return
	}

	notification := &domain.Notification{
		Title:      req.Title,
		Body:       req.Body,
		TargetType: req.Target,
	}

	if err := h.notificationService.SendToAll(c.Request.Context(), notification); err != nil {
		response.InternalError(c, "failed to send notification")
		return
	}

	// Persist to notification history
	adminUserID, _ := c.Get(middleware.ContextKeyUserID)
	adminID := adminUserID.(uuid.UUID)
	history := &domain.NotificationHistory{
		ID:          uuid.New(),
		AdminUserID: adminID,
		Title:       req.Title,
		Body:        req.Body,
		TargetType:  req.Target,
	}
	_ = h.historyRepo.Create(c.Request.Context(), history)

	response.OK(c, gin.H{"message": "notification sent successfully"})
}
