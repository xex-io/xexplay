package admin

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

type RewardHandler struct {
	rewardService *service.RewardService
}

func NewRewardHandler(rewardService *service.RewardService) *RewardHandler {
	return &RewardHandler{rewardService: rewardService}
}

// ListConfigs handles GET /admin/rewards/configs
func (h *RewardHandler) ListConfigs(c *gin.Context) {
	configs, err := h.rewardService.ListAllConfigs(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to list reward configs")
		return
	}
	response.OK(c, configs)
}

type createConfigRequest struct {
	PeriodType  string          `json:"period_type" binding:"required"`
	RankFrom    int             `json:"rank_from" binding:"required,min=1"`
	RankTo      int             `json:"rank_to" binding:"required,min=1"`
	RewardType  string          `json:"reward_type" binding:"required"`
	Amount      float64         `json:"amount" binding:"min=0"`
	Description json.RawMessage `json:"description"`
	IsActive    *bool           `json:"is_active"`
}

// CreateConfig handles POST /admin/rewards/configs
func (h *RewardHandler) CreateConfig(c *gin.Context) {
	var req createConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	cfg := &domain.RewardConfig{
		ID:          uuid.New(),
		PeriodType:  req.PeriodType,
		RankFrom:    req.RankFrom,
		RankTo:      req.RankTo,
		RewardType:  req.RewardType,
		Amount:      req.Amount,
		Description: req.Description,
		IsActive:    isActive,
	}

	if err := h.rewardService.CreateConfig(c.Request.Context(), cfg); err != nil {
		response.InternalError(c, "failed to create reward config")
		return
	}

	response.Created(c, cfg)
}

type updateConfigRequest struct {
	PeriodType  string          `json:"period_type" binding:"required"`
	RankFrom    int             `json:"rank_from" binding:"required,min=1"`
	RankTo      int             `json:"rank_to" binding:"required,min=1"`
	RewardType  string          `json:"reward_type" binding:"required"`
	Amount      float64         `json:"amount" binding:"min=0"`
	Description json.RawMessage `json:"description"`
	IsActive    *bool           `json:"is_active"`
}

// UpdateConfig handles PUT /admin/rewards/configs/:id
func (h *RewardHandler) UpdateConfig(c *gin.Context) {
	idParam := c.Param("id")
	cfgID, err := uuid.Parse(idParam)
	if err != nil {
		response.BadRequest(c, "invalid config id")
		return
	}

	var req updateConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	cfg := &domain.RewardConfig{
		ID:          cfgID,
		PeriodType:  req.PeriodType,
		RankFrom:    req.RankFrom,
		RankTo:      req.RankTo,
		RewardType:  req.RewardType,
		Amount:      req.Amount,
		Description: req.Description,
		IsActive:    isActive,
	}

	if err := h.rewardService.UpdateConfig(c.Request.Context(), cfg); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, cfg)
}

type distributeRequest struct {
	PeriodType string                   `json:"period_type" binding:"required"`
	PeriodKey  string                   `json:"period_key" binding:"required"`
	Entries    []domain.RewardLeaderboardEntry `json:"entries" binding:"required"`
}

// Distribute handles POST /admin/rewards/distribute
func (h *RewardHandler) Distribute(c *gin.Context) {
	var req distributeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	count, err := h.rewardService.DistributeRewards(c.Request.Context(), req.PeriodType, req.PeriodKey, req.Entries)
	if err != nil {
		response.InternalError(c, "failed to distribute rewards")
		return
	}

	response.OK(c, gin.H{
		"distributed": count,
	})
}

// GetHistory handles GET /admin/rewards/history
func (h *RewardHandler) GetHistory(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	history, err := h.rewardService.GetDistributionHistory(c.Request.Context(), limit, offset)
	if err != nil {
		response.InternalError(c, "failed to get distribution history")
		return
	}

	response.OK(c, history)
}
