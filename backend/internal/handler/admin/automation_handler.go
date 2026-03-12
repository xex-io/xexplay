package admin

import (
	"github.com/gin-gonic/gin"

	"github.com/xex-exchange/xexplay-api/internal/pkg/response"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

type AutomationHandler struct {
	sportRepo     *postgres.SportRepo
	logRepo       *postgres.AutomationLogRepo
	sportsDataSvc *service.SportsDataService
	autoResolveSvc *service.AutoResolveService
}

func NewAutomationHandler(
	sportRepo *postgres.SportRepo,
	logRepo *postgres.AutomationLogRepo,
	sportsDataSvc *service.SportsDataService,
	autoResolveSvc *service.AutoResolveService,
) *AutomationHandler {
	return &AutomationHandler{
		sportRepo:      sportRepo,
		logRepo:        logRepo,
		sportsDataSvc:  sportsDataSvc,
		autoResolveSvc: autoResolveSvc,
	}
}

// ListSports handles GET /admin/sports
func (h *AutomationHandler) ListSports(c *gin.Context) {
	sports, err := h.sportRepo.FindAll(c.Request.Context())
	if err != nil {
		response.InternalError(c, "failed to fetch sports")
		return
	}
	response.OK(c, sports)
}

type toggleSportRequest struct {
	IsActive bool `json:"is_active"`
}

// ToggleSport handles PUT /admin/sports/:key
func (h *AutomationHandler) ToggleSport(c *gin.Context) {
	key := c.Param("key")

	var req toggleSportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if err := h.sportRepo.SetActive(c.Request.Context(), key, req.IsActive); err != nil {
		response.InternalError(c, "failed to update sport: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "sport updated", "key": key, "is_active": req.IsActive})
}

// GetAutomationLogs handles GET /admin/automation/logs
func (h *AutomationHandler) GetAutomationLogs(c *gin.Context) {
	jobName := c.Query("job")
	limit := 50

	var logs interface{}
	var err error
	if jobName != "" {
		logs, err = h.logRepo.FindByJob(c.Request.Context(), jobName, limit)
	} else {
		logs, err = h.logRepo.FindRecent(c.Request.Context(), limit)
	}

	if err != nil {
		response.InternalError(c, "failed to fetch automation logs")
		return
	}
	response.OK(c, logs)
}

type triggerJobRequest struct {
	Job string `json:"job" binding:"required"`
}

// TriggerJob handles POST /admin/automation/trigger
func (h *AutomationHandler) TriggerJob(c *gin.Context) {
	var req triggerJobRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body: job field is required")
		return
	}

	ctx := c.Request.Context()
	var err error

	switch req.Job {
	case "fetchMatches":
		err = h.sportsDataSvc.FetchUpcomingMatches(ctx)
	case "generateCards":
		err = h.sportsDataSvc.GenerateDailyCards(ctx)
	case "autoPublish":
		err = h.sportsDataSvc.AutoPublishBaskets(ctx)
	case "autoResolve":
		err = h.autoResolveSvc.ProcessCompletedMatches(ctx)
	case "syncSports":
		err = h.sportsDataSvc.SyncSports(ctx)
	default:
		response.BadRequest(c, "unknown job: "+req.Job+". Valid jobs: fetchMatches, generateCards, autoPublish, autoResolve, syncSports")
		return
	}

	if err != nil {
		response.InternalError(c, "job failed: "+err.Error())
		return
	}

	response.OK(c, gin.H{"message": "job completed", "job": req.Job})
}

// GetAutomationStatus handles GET /admin/automation/status
func (h *AutomationHandler) GetAutomationStatus(c *gin.Context) {
	ctx := c.Request.Context()

	jobs := []string{"fetchUpcomingMatches", "generateDailyCards", "autoPublishBaskets", "autoResolveCards", "syncSports"}
	status := make(map[string]interface{})

	for _, job := range jobs {
		logs, err := h.logRepo.FindByJob(ctx, job, 1)
		if err != nil || len(logs) == 0 {
			status[job] = gin.H{"last_run": nil, "status": "never"}
		} else {
			status[job] = gin.H{
				"last_run":        logs[0].CreatedAt,
				"status":          logs[0].Status,
				"items_processed": logs[0].ItemsProcessed,
				"error":           logs[0].ErrorMessage,
			}
		}
	}

	response.OK(c, status)
}
