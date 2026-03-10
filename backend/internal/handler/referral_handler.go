package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

type ReferralHandler struct {
	referralService *service.ReferralService
}

func NewReferralHandler(referralService *service.ReferralService) *ReferralHandler {
	return &ReferralHandler{referralService: referralService}
}

// GetReferralCode handles GET /referral/code — returns the user's referral code.
func (h *ReferralHandler) GetReferralCode(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	code, err := h.referralService.GetReferralCode(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, "failed to get referral code")
		return
	}

	response.OK(c, gin.H{
		"referral_code": code,
	})
}

// GetReferralStats handles GET /referral/stats — returns referral statistics for the user.
func (h *ReferralHandler) GetReferralStats(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	stats, err := h.referralService.GetReferralStats(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, "failed to get referral stats")
		return
	}

	response.OK(c, stats)
}
