package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

type AchievementHandler struct {
	achievementService *service.AchievementService
}

func NewAchievementHandler(achievementService *service.AchievementService) *AchievementHandler {
	return &AchievementHandler{achievementService: achievementService}
}

// GetMyAchievements handles GET /me/achievements — returns all achievements with user's earned status.
func (h *AchievementHandler) GetMyAchievements(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	allAchievements, err := h.achievementService.GetAllAchievements(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to get achievements")
		return
	}

	userAchievements, err := h.achievementService.GetUserAchievements(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, "failed to get user achievements")
		return
	}

	response.OK(c, gin.H{
		"achievements": allAchievements,
		"earned":       userAchievements,
	})
}
