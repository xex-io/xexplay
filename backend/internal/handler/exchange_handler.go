package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

// ExchangeHandler handles Exchange integration endpoints.
// These are stub endpoints that will be connected to the real Exchange API in the future.
type ExchangeHandler struct {
	rewardService *service.RewardService
	userRepo      *postgres.UserRepo
}

func NewExchangeHandler(rewardService *service.RewardService, userRepo *postgres.UserRepo) *ExchangeHandler {
	return &ExchangeHandler{rewardService: rewardService, userRepo: userRepo}
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
	// This includes Exchange account verification for token and trading_fee_discount rewards.
	if err := h.rewardService.ClaimReward(c.Request.Context(), distID, uid); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Step 2: Determine credit action based on reward type
	dist, err := h.rewardService.GetDistributionByID(c.Request.Context(), distID)
	if err != nil {
		response.InternalError(c, "reward claimed but failed to retrieve details")
		return
	}

	switch dist.RewardType {
	case domain.RewardTradingFeeDiscount:
		// TODO: In production, this will call the Exchange API:
		//   POST /api/v1/trading/fee-discount { user_id, discount_percent, duration_days, reference: distID }
		// For now, we mark it as credited directly.
		if err := h.rewardService.CreditReward(c.Request.Context(), distID, uid); err != nil {
			response.InternalError(c, "reward claimed but trading fee discount activation failed — will retry")
			return
		}
		response.OK(c, gin.H{
			"message":     "trading fee discount activated on exchange",
			"status":      "credited",
			"reward_type": domain.RewardTradingFeeDiscount,
			"amount":      dist.Amount,
		})

	case domain.RewardToken:
		// TODO: In production, this will call the Exchange API:
		//   POST /api/v1/wallets/credit { user_id, amount, currency: "XEX", reference: distID }
		// For now, we mark it as credited directly.
		if err := h.rewardService.CreditReward(c.Request.Context(), distID, uid); err != nil {
			response.InternalError(c, "reward claimed but credit to exchange failed — will retry")
			return
		}
		response.OK(c, gin.H{
			"message":     "reward credited to exchange wallet",
			"status":      "credited",
			"reward_type": domain.RewardToken,
			"amount":      dist.Amount,
		})

	default:
		// Non-exchange reward types (badge, bonus_skip, bonus_answer) are claimed directly.
		response.OK(c, gin.H{
			"message":     "reward claimed successfully",
			"status":      "claimed",
			"reward_type": dist.RewardType,
		})
	}
}

// GetExchangePrompts handles GET /me/exchange-prompts
// Returns contextual Exchange prompts based on user state.
// These prompts encourage users to visit XEX Exchange for trading, staking, etc.
func (h *ExchangeHandler) GetExchangePrompts(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)

	// Gather user and reward state to generate contextual prompts
	user, err := h.userRepo.FindByID(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, "failed to get user state")
		return
	}

	pending, err := h.rewardService.GetPendingRewards(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, "failed to get reward state")
		return
	}

	prompts := []gin.H{}

	// Prompt: unclaimed rewards
	if len(pending) > 0 {
		var totalTokens float64
		var hasFeeDiscount bool
		for _, r := range pending {
			if r.RewardType == domain.RewardToken {
				totalTokens += r.Amount
			}
			if r.RewardType == domain.RewardTradingFeeDiscount {
				hasFeeDiscount = true
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
		if hasFeeDiscount {
			prompts = append(prompts, gin.H{
				"type":    "unclaimed_fee_discount",
				"title":   "Trading fee discount available!",
				"message": "You earned a trading fee discount. Claim it and save on your next trades.",
				"cta_url": "https://xex.exchange/trade",
			})
		}
	}

	// Prompt: active trader VIP cards
	if user != nil && user.IsActiveTrader() {
		prompts = append(prompts, gin.H{
			"type":         "vip_trader",
			"title":        "VIP Trader Cards Unlocked!",
			"message":      "As an active Exchange trader, you have access to exclusive VIP prediction cards with higher point rewards.",
			"trading_tier": user.TradingTier,
		})
	}

	// Prompt: Exchange account linking
	if user != nil && user.ExchangeStatus == "" {
		prompts = append(prompts, gin.H{
			"type":    "link_exchange",
			"title":   "Link your Exchange account",
			"message": "Link your XEX Exchange account to unlock token rewards, trading fee discounts, and VIP cards.",
			"cta_url": "https://xex.exchange/settings/linked-apps",
		})
	}

	// Prompt: general exchange onboarding (only for non-traders)
	if user == nil || !user.IsActiveTrader() {
		prompts = append(prompts, gin.H{
			"type":    "exchange_onboarding",
			"title":   "Trade on XEX Exchange",
			"message": "Turn your prediction skills into real trades. Start trading on XEX Exchange today.",
			"cta_url": "https://xex.exchange/trade",
		})
	}

	response.OK(c, gin.H{
		"prompts": prompts,
	})
}
