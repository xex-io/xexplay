package admin

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

type AuditHandler struct {
	auditService *service.AuditService
	abuseService *service.AbuseService
}

func NewAuditHandler(auditService *service.AuditService, abuseService *service.AbuseService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
		abuseService: abuseService,
	}
}

// GetAuditLogs handles GET /admin/audit-logs — paginated audit log viewer.
func (h *AuditHandler) GetAuditLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Optional filter by admin user
	adminIDParam := c.Query("admin_user_id")
	if adminIDParam != "" {
		adminID, err := uuid.Parse(adminIDParam)
		if err != nil {
			response.BadRequest(c, "invalid admin_user_id")
			return
		}
		logs, err := h.auditService.GetLogsByAdmin(c.Request.Context(), adminID, limit, offset)
		if err != nil {
			response.InternalError(c, "failed to get audit logs")
			return
		}
		response.OK(c, logs)
		return
	}

	logs, err := h.auditService.GetRecentLogs(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c, "failed to get audit logs")
		return
	}
	response.OK(c, logs)
}

// GetAbuseFlags handles GET /admin/abuse-flags — pending abuse flags for review.
func (h *AuditHandler) GetAbuseFlags(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	flags, err := h.abuseService.GetPendingFlags(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c, "failed to get abuse flags")
		return
	}
	response.OK(c, flags)
}

type reviewFlagRequest struct {
	Status string `json:"status" binding:"required"`
}

// ReviewAbuseFlag handles POST /admin/abuse-flags/:id/review — review a flag (approve/dismiss).
func (h *AuditHandler) ReviewAbuseFlag(c *gin.Context) {
	flagIDParam := c.Param("id")
	flagID, err := uuid.Parse(flagIDParam)
	if err != nil {
		response.BadRequest(c, "invalid flag id")
		return
	}

	var req reviewFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	adminUserID, _ := c.Get("user_id")
	adminID := adminUserID.(uuid.UUID)

	if err := h.abuseService.ReviewFlag(c.Request.Context(), flagID, adminID, req.Status); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "flag reviewed successfully"})
}
