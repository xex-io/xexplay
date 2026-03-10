package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

// ExchangeHandler handles Exchange integration endpoints.
// These are stub endpoints that will be connected to the real Exchange API in the future.
type ExchangeHandler struct {
	rewardService *service.RewardService
}

func NewExchangeHandler(rewardService *service.RewardService) *ExchangeHandler {
	return &ExchangeHandler{rewardService: rewardService}
}

// ClaimRewardToExchange handles POST /me/rewards/:id/claim
// This is the exchange-integration version: it claims the reward and sets status to 'credited'.
// In future, this will call the Exchange API to credit tokens to the user's Exchange wallet.
func (h *ExchangeHandler) ClaimRewardToExchange(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	idParam := c.Param("id")
	distID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid reward id")
		return
	}

	// Step 1: Claim the reward (pending -> claimed)
	if err := h.rewardService.ClaimReward(c.Request.Context(), distID, uid); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Step 2: Credit the reward (claimed -> credited)
	// TODO: In production, this will call the Exchange API:
	//   POST /api/v1/wallets/credit { user_id, amount, currency: "XEX", reference: distID }
	// For now, we mark it as credited directly.
	if err := h.rewardService.CreditReward(c.Request.Context(), distID, uid); err != nil {
		response.InternalError(c, "reward claimed but credit to exchange failed — will retry")
		return
	}

	response.OK(c, gin.H{
		"message": "reward credited to exchange wallet",
		"status":  "credited",
	})
}

// GetExchangePrompts handles GET /me/exchange-prompts
// Returns contextual Exchange prompts based on user state.
// These prompts encourage users to visit XEX Exchange for trading, staking, etc.
func (h *ExchangeHandler) GetExchangePrompts(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	// Gather user reward state to generate contextual prompts
	pending, err := h.rewardService.GetPendingRewards(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, "failed to get reward state")
		return
	}

	prompts := []gin.H{}

	// Prompt: unclaimed rewards
	if len(pending) > 0 {
		var totalTokens float64
		for _, r := range pending {
			if r.RewardType == "token" {
				totalTokens += r.Amount
			}
		}
		if totalTokens > 0 {
			prompts = append(prompts, gin.H{
				"type":    "unclaimed_tokens",
				"title":   "Claim your XEX tokens!",
				"message": "You have unclaimed token rewards. Claim them and trade on XEX Exchange.",
				"amount":  totalTokens,
				"cta_url": "https://xex.exchange/wallet",
			})
		}
	}

	// Prompt: general exchange onboarding
	prompts = append(prompts, gin.H{
		"type":    "exchange_onboarding",
		"title":   "Trade on XEX Exchange",
		"message": "Turn your prediction skills into real trades. Start trading on XEX Exchange today.",
		"cta_url": "https://xex.exchange/trade",
	})

	response.OK(c, gin.H{
		"prompts": prompts,
	})
}
