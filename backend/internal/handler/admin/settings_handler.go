package admin

import (
	"github.com/gin-gonic/gin"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

type SettingsHandler struct {
	settingRepo *postgres.SettingRepo
}

func NewSettingsHandler(settingRepo *postgres.SettingRepo) *SettingsHandler {
	return &SettingsHandler{settingRepo: settingRepo}
}

// List handles GET /admin/settings
func (h *SettingsHandler) List(c *gin.Context) {
	settings, err := h.settingRepo.FindAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to fetch settings")
		return
	}

	// Return masked views for secret values
	views := make([]interface{}, 0, len(settings))
	for _, s := range settings {
		views = append(views, s.ToView())
	}
	response.OK(c, views)
}

type updateSettingRequest struct {
	Value string `json:"value"`
}

// Update handles PUT /admin/settings/:key
func (h *SettingsHandler) Update(c *gin.Context) {
	key := c.Param("key")

	var req updateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.settingRepo.Set(c.Request.Context(), key, req.Value); err != nil {
		response.InternalError(c, "failed to update setting: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "setting updated", "key": key})
}

// Delete handles DELETE /admin/settings/:key (clears the value)
func (h *SettingsHandler) Delete(c *gin.Context) {
	key := c.Param("key")

	if err := h.settingRepo.Delete(c.Request.Context(), key); err != nil {
		response.InternalError(c, "failed to clear setting: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "setting cleared", "key": key})
}
