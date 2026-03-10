package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

type RewardHandler struct {
	rewardService *service.RewardService
	streakService *service.StreakService
}

func NewRewardHandler(rewardService *service.RewardService, streakService *service.StreakService) *RewardHandler {
	return &RewardHandler{
		rewardService: rewardService,
		streakService: streakService,
	}
}

// GetRewards handles GET /me/rewards — returns pending and historical rewards.
func (h *RewardHandler) GetRewards(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	pending, err := h.rewardService.GetPendingRewards(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, "failed to get pending rewards")
		return
	}

	history, err := h.rewardService.GetRewardHistory(c.Request.Context(), uid, limit, offset)
	if err != nil {
		response.InternalError(c, "failed to get reward history")
		return
	}

	streak, err := h.streakService.GetStreak(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, "failed to get streak info")
		return
	}

	response.OK(c, gin.H{
		"pending": pending,
		"history": history,
		"streak":  streak,
	})
}

// ClaimReward handles POST /me/rewards/:id/claim — claims a pending reward.
func (h *RewardHandler) ClaimReward(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	idParam := c.Param("id")
	distID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid reward id")
		return
	}

	if err := h.rewardService.ClaimReward(c.Request.Context(), distID, uid); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "reward claimed successfully"})
}
