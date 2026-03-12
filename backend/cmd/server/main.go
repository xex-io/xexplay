package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xex-exchange/xexplay-api/internal/config"
	"github.com/xex-exchange/xexplay-api/internal/external/oddsapi"
	"github.com/xex-exchange/xexplay-api/internal/handler"
	adminHandler "github.com/xex-exchange/xexplay-api/internal/handler/admin"
	wsHandler "github.com/xex-exchange/xexplay-api/internal/handler/ws"
	"github.com/xex-exchange/xexplay-api/internal/middleware"
	"github.com/xex-exchange/xexplay-api/internal/pkg/ws"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
	"github.com/xex-exchange/xexplay-api/internal/repository/redis"
	"github.com/xex-exchange/xexplay-api/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	// Setup logging
	setupLogging(cfg.LogLevel)

	// Connect to PostgreSQL
	db, err := postgres.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to PostgreSQL")
	}
	defer db.Close()

	// Run migrations
	if err := postgres.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}

	// Connect to Redis
	rdb, err := redis.NewConnection(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to Redis")
	}
	defer rdb.Close()

	// Initialize WebSocket hub
	wsHub := ws.NewHub()
	go wsHub.Run()

	// Initialize repositories
	repos := initRepositories(db, rdb)

	// Initialize services
	services := initServices(cfg, repos, wsHub)

	// Initialize handlers
	handlers := initHandlers(cfg, services, repos, wsHub)

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// Global middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS(cfg.CORSOrigins))
	router.Use(middleware.Metrics())

	// Health check endpoints (no auth required)
	healthHandler := handler.NewHealthHandler(db, rdb)
	router.GET("/health", healthHandler.Liveness)
	router.GET("/health/ready", healthHandler.Readiness)

	// Prometheus metrics endpoint (no auth required)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Public routes
	public := router.Group("/v1")
	{
		public.POST("/auth/login", handlers.Auth.Login)
	}

	// Authenticated routes
	authed := router.Group("/v1")
	authed.Use(middleware.Auth(cfg.JWTSecret, repos.User))
	authed.Use(middleware.Locale())
	authed.Use(middleware.RateLimiter(rdb, 30, time.Minute))
	{
		// User profile
		authed.GET("/me", handlers.User.GetProfile)
		authed.PUT("/me", handlers.User.UpdateProfile)
		authed.GET("/me/stats", handlers.User.GetStats)
		authed.GET("/me/history", handlers.User.GetHistory)

		// Game session
		authed.POST("/sessions/start", handlers.Game.StartSession)
		authed.GET("/sessions/current", handlers.Game.GetCurrentSession)
		authed.GET("/sessions/current/card", handlers.Game.GetCurrentCard)
		authed.POST("/sessions/current/answer", handlers.Game.SubmitAnswer)
		authed.POST("/sessions/current/skip", handlers.Game.SkipCard)

		// Rewards
		authed.GET("/me/rewards", handlers.Reward.GetRewards)
		authed.POST("/me/rewards/:id/claim", handlers.Exchange.ClaimRewardToExchange)

		// Exchange integration
		authed.GET("/me/exchange-prompts", handlers.Exchange.GetExchangePrompts)

		// Events
		authed.GET("/events", handlers.Event.List)
		authed.GET("/events/:id", handlers.Event.GetByID)

		// Leaderboards
		authed.GET("/leaderboards/daily", handlers.Leaderboard.GetDaily)
		authed.GET("/leaderboards/weekly", handlers.Leaderboard.GetWeekly)
		authed.GET("/leaderboards/tournament/:eventId", handlers.Leaderboard.GetTournament)
		authed.GET("/leaderboards/all-time", handlers.Leaderboard.GetAllTime)
		authed.GET("/leaderboards/friends", handlers.Friends.GetFriendsLeaderboard)

		// Achievements
		authed.GET("/me/achievements", handlers.Achievement.GetMyAchievements)

		// Referrals
		authed.GET("/referral/code", handlers.Referral.GetReferralCode)
		authed.GET("/referral/stats", handlers.Referral.GetReferralStats)

		// Mini Leagues
		authed.POST("/leagues", handlers.League.CreateLeague)
		authed.POST("/leagues/join", handlers.League.JoinLeague)
		authed.GET("/leagues", handlers.League.GetMyLeagues)
		authed.GET("/leagues/:id", handlers.League.GetLeague)
		authed.GET("/leaderboards/league/:leagueId", handlers.League.GetLeagueLeaderboard)

		// Devices (push notification tokens)
		authed.POST("/devices/register", handlers.Device.Register)
		authed.DELETE("/devices/:token", handlers.Device.Deregister)
	}

	// WebSocket (authenticated via token query param)
	router.GET("/ws", handlers.WebSocket.Handle)

	// Admin routes (authenticated via Exchange admin session)
	admin := router.Group("/v1/admin")
	admin.Use(middleware.ExchangeAdminAuth(cfg.ExchangeAPIURL))
	admin.Use(middleware.RateLimiter(rdb, 100, time.Minute))
	admin.Use(middleware.AuditLog(services.Audit))
	{
		// Events
		admin.GET("/events", handlers.AdminEvent.List)
		admin.POST("/events", handlers.AdminEvent.Create)
		admin.PUT("/events/:id", handlers.AdminEvent.Update)
		admin.DELETE("/events/:id", handlers.AdminEvent.Delete)

		// Matches
		admin.GET("/matches", handlers.AdminMatch.List)
		admin.POST("/matches", handlers.AdminMatch.Create)
		admin.PUT("/matches/:id", handlers.AdminMatch.Update)
		admin.DELETE("/matches/:id", handlers.AdminMatch.Delete)

		// Cards
		admin.GET("/cards", handlers.AdminCard.List)
		admin.POST("/cards", handlers.AdminCard.Create)
		admin.PUT("/cards/:id", handlers.AdminCard.Update)
		admin.DELETE("/cards/:id", handlers.AdminCard.Delete)
		admin.POST("/cards/:id/resolve", handlers.AdminCard.Resolve)

		// Baskets
		admin.GET("/baskets", handlers.AdminBasket.List)
		admin.POST("/baskets", handlers.AdminBasket.Create)
		admin.PUT("/baskets/:id", handlers.AdminBasket.Update)
		admin.DELETE("/baskets/:id", handlers.AdminBasket.Delete)
		admin.POST("/baskets/:id/publish", handlers.AdminBasket.Publish)

		// Users
		admin.GET("/users", handlers.AdminUser.List)
		admin.GET("/users/:id", handlers.AdminUser.GetByID)
		admin.PUT("/users/:id", handlers.AdminUser.Update)

		// Rewards
		admin.GET("/rewards/configs", handlers.AdminReward.ListConfigs)
		admin.POST("/rewards/configs", handlers.AdminReward.CreateConfig)
		admin.PUT("/rewards/configs/:id", handlers.AdminReward.UpdateConfig)
		admin.POST("/rewards/distribute", handlers.AdminReward.Distribute)
		admin.GET("/rewards/history", handlers.AdminReward.GetHistory)

		// Notifications
		admin.POST("/notifications/send", handlers.AdminNotification.Send)

		// Audit logs
		admin.GET("/audit-logs", handlers.AdminAudit.GetAuditLogs)

		// Abuse flags
		admin.GET("/abuse-flags", handlers.AdminAudit.GetAbuseFlags)
		admin.POST("/abuse-flags/:id/review", handlers.AdminAudit.ReviewAbuseFlag)
		admin.GET("/abuse-flags/stats", handlers.AdminDashboard.AdminGetAbuseFlagStats)

		// Leaderboards (admin view)
		admin.GET("/leaderboards/:type", handlers.AdminDashboard.AdminGetLeaderboard)

		// Analytics
		admin.GET("/analytics/overview", handlers.AdminDashboard.AdminGetAnalytics)

		// Exchange metrics
		admin.GET("/exchange/metrics", handlers.AdminDashboard.AdminGetExchangeMetrics)

		// User search and activity
		admin.GET("/users/search", handlers.AdminDashboard.AdminSearchUsers)
		admin.GET("/users/:id/activity", handlers.AdminDashboard.AdminGetUserActivity)
		admin.POST("/users/:id/moderate", handlers.AdminDashboard.AdminModerateUser)

		// Notification history
		admin.GET("/notifications", handlers.AdminDashboard.AdminListNotifications)

		// Referral stats
		admin.GET("/referrals/stats", handlers.AdminDashboard.AdminGetReferralStats)
		admin.GET("/referrals/top", handlers.AdminDashboard.AdminGetTopReferrers)

		// Prize pools
		admin.GET("/prize-pools", handlers.AdminDashboard.AdminListPrizePools)
		admin.GET("/prize-pools/history", handlers.AdminDashboard.AdminGetPrizePoolHistory)
		admin.POST("/prize-pools", handlers.AdminDashboard.AdminCreatePrizePool)
		admin.PUT("/prize-pools/:id", handlers.AdminDashboard.AdminUpdatePrizePool)
		admin.DELETE("/prize-pools/:id", handlers.AdminDashboard.AdminCancelPrizePool)

		// Sports automation
		admin.GET("/sports", handlers.AdminAutomation.ListSports)
		admin.PUT("/sports/:key", handlers.AdminAutomation.ToggleSport)
		admin.GET("/automation/status", handlers.AdminAutomation.GetAutomationStatus)
		admin.POST("/automation/trigger", handlers.AdminAutomation.TriggerJob)
		admin.GET("/automation/logs", handlers.AdminAutomation.GetAutomationLogs)

		// Settings
		admin.GET("/settings", handlers.AdminSettings.List)
		admin.PUT("/settings/:key", handlers.AdminSettings.Update)
		admin.DELETE("/settings/:key", handlers.AdminSettings.Delete)
	}

	// Start cron jobs
	cronCtx, cronCancel := context.WithCancel(context.Background())
	defer cronCancel()
	services.Cron.StartCronJobs(cronCtx)

	// Start card expiry monitor
	cardExpiryMonitor := service.NewCardExpiryMonitor(repos.Card, wsHub)
	go cardExpiryMonitor.Start(cronCtx)

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("port", cfg.Port).Msg("starting XEX Play API server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutting down server...")

	// Stop cron jobs
	cronCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server exited cleanly")
}

func setupLogging(level string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
}

type repositories struct {
	User                *postgres.UserRepo
	Event               *postgres.EventRepo
	Match               *postgres.MatchRepo
	Card                *postgres.CardRepo
	Basket              *postgres.BasketRepo
	Session             *postgres.SessionRepo
	Answer              *postgres.AnswerRepo
	Leaderboard         *postgres.LeaderboardRepo
	Streak              *postgres.StreakRepo
	Reward              *postgres.RewardRepo
	FCMToken            *postgres.FCMTokenRepo
	Achievement         *postgres.AchievementRepo
	Referral            *postgres.ReferralRepo
	MiniLeague          *postgres.MiniLeagueRepo
	Audit               *postgres.AuditRepo
	Abuse               *postgres.AbuseRepo
	NotificationHistory *postgres.NotificationHistoryRepo
	PrizePool           *postgres.PrizePoolRepo
	Sport               *postgres.SportRepo
	AutomationLog       *postgres.AutomationLogRepo
	Setting             *postgres.SettingRepo
	Cache               *redis.CacheRepo
	LBCache             *redis.LeaderboardCache
}

func initRepositories(db *postgres.DB, rdb *redis.Client) *repositories {
	return &repositories{
		User:                postgres.NewUserRepo(db),
		Event:               postgres.NewEventRepo(db),
		Match:               postgres.NewMatchRepo(db),
		Card:                postgres.NewCardRepo(db),
		Basket:              postgres.NewBasketRepo(db),
		Session:             postgres.NewSessionRepo(db),
		Answer:              postgres.NewAnswerRepo(db),
		Leaderboard:         postgres.NewLeaderboardRepo(db),
		Streak:              postgres.NewStreakRepo(db),
		Reward:              postgres.NewRewardRepo(db),
		FCMToken:            postgres.NewFCMTokenRepo(db),
		Achievement:         postgres.NewAchievementRepo(db),
		Referral:            postgres.NewReferralRepo(db),
		MiniLeague:          postgres.NewMiniLeagueRepo(db),
		Audit:               postgres.NewAuditRepo(db),
		Abuse:               postgres.NewAbuseRepo(db),
		NotificationHistory: postgres.NewNotificationHistoryRepo(db),
		PrizePool:           postgres.NewPrizePoolRepo(db),
		Sport:               postgres.NewSportRepo(db),
		AutomationLog:       postgres.NewAutomationLogRepo(db),
		Setting:             postgres.NewSettingRepo(db),
		Cache:               redis.NewCacheRepo(rdb),
		LBCache:             redis.NewLeaderboardCache(rdb),
	}
}

type services struct {
	Auth           *service.AuthService
	Game           *service.GameService
	Card           *service.CardService
	Shuffle        *service.ShuffleService
	Leaderboard    *service.LeaderboardService
	Streak         *service.StreakService
	Reward         *service.RewardService
	Notification   *service.NotificationService
	Cron           *service.CronService
	Achievement    *service.AchievementService
	Referral       *service.ReferralService
	MiniLeague     *service.MiniLeagueService
	Audit          *service.AuditService
	Abuse          *service.AbuseService
	SportsData     *service.SportsDataService
	AutoResolve    *service.AutoResolveService
	AI             *service.AIService
}

func initServices(cfg *config.Config, repos *repositories, wsHub *ws.Hub) *services {
	shuffleSvc := service.NewShuffleService()
	leaderboardSvc := service.NewLeaderboardService(repos.Leaderboard, repos.LBCache, repos.User, wsHub)
	streakSvc := service.NewStreakService(repos.Streak)
	rewardSvc := service.NewRewardService(repos.Reward, repos.User, wsHub)

	// Notification service: use FCM in production, log sender in development
	var notifSender service.NotificationSender
	if cfg.FCMCredentialsJSON != "" {
		fcmSender, err := service.NewFCMSenderFromJSON(cfg.FCMCredentialsJSON)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to initialize FCM sender from JSON")
		}
		notifSender = fcmSender
		log.Info().Msg("using FCM sender for push notifications (from JSON)")
	} else if cfg.FCMCredentialsFile != "" {
		fcmSender, err := service.NewFCMSender(cfg.FCMCredentialsFile)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to initialize FCM sender")
		}
		notifSender = fcmSender
		log.Info().Msg("using FCM sender for push notifications (from file)")
	} else {
		notifSender = service.NewLogSender()
		log.Info().Msg("using log sender for push notifications (set FCM_CREDENTIALS_JSON for production)")
	}
	notificationSvc := service.NewNotificationService(repos.FCMToken, notifSender)

	// Cron service
	cronSvc := service.NewCronService(
		repos.Leaderboard,
		leaderboardSvc,
		rewardSvc,
		repos.Streak,
		notificationSvc,
		repos.LBCache,
	)

	achievementSvc := service.NewAchievementService(repos.Achievement, wsHub)
	referralSvc := service.NewReferralService(repos.Referral, repos.User)
	miniLeagueSvc := service.NewMiniLeagueService(repos.MiniLeague, repos.Leaderboard, repos.User, repos.LBCache)
	auditSvc := service.NewAuditService(repos.Audit)
	abuseSvc := service.NewAbuseService(repos.Abuse, repos.User, repos.Reward)

	cardSvc := service.NewCardService(repos.Card, repos.Answer, leaderboardSvc, wsHub)

	// Sports automation services — load API keys from DB first, fallback to env vars
	var aiSvc *service.AIService
	var sportsDataSvc *service.SportsDataService
	var autoResolveSvc *service.AutoResolveService

	oddsAPIKey := cfg.OddsAPIKey
	anthropicAPIKey := cfg.AnthropicAPIKey
	autoSportsEnabled := cfg.AutoSportsEnabled

	// Try loading from DB settings
	if dbKey, err := repos.Setting.Get(context.Background(), "ODDS_API_KEY"); err == nil && dbKey != "" {
		oddsAPIKey = dbKey
	}
	if dbKey, err := repos.Setting.Get(context.Background(), "ANTHROPIC_API_KEY"); err == nil && dbKey != "" {
		anthropicAPIKey = dbKey
	}
	if dbVal, err := repos.Setting.Get(context.Background(), "AUTO_SPORTS_ENABLED"); err == nil && dbVal != "" {
		autoSportsEnabled = dbVal == "true"
	}

	if oddsAPIKey != "" && anthropicAPIKey != "" {
		oddsClient := oddsapi.NewClient(oddsAPIKey)
		aiSvc = service.NewAIService(anthropicAPIKey)

		sportsDataSvc = service.NewSportsDataService(
			oddsClient, repos.Match, repos.Event, repos.Card, repos.Basket,
			repos.Sport, aiSvc, repos.AutomationLog,
		)
		autoResolveSvc = service.NewAutoResolveService(
			repos.Match, repos.Card, cardSvc, oddsClient, aiSvc, repos.AutomationLog,
		)

		cronSvc.SetAutomationServices(sportsDataSvc, autoResolveSvc, autoSportsEnabled)
		log.Info().Msg("sports automation services initialized")
	} else {
		log.Info().Msg("sports automation disabled (ODDS_API_KEY or ANTHROPIC_API_KEY not set)")
	}

	return &services{
		Auth:         service.NewAuthService(repos.User, referralSvc, cfg.JWTSecret),
		Game:         service.NewGameService(repos.Session, repos.Answer, repos.Basket, repos.Card, repos.User, repos.Cache, shuffleSvc, streakSvc, achievementSvc),
		Card:         cardSvc,
		Shuffle:      shuffleSvc,
		Leaderboard:  leaderboardSvc,
		Streak:       streakSvc,
		Reward:       rewardSvc,
		Notification: notificationSvc,
		Cron:         cronSvc,
		Achievement:  achievementSvc,
		Referral:     referralSvc,
		MiniLeague:   miniLeagueSvc,
		Audit:        auditSvc,
		Abuse:        abuseSvc,
		SportsData:   sportsDataSvc,
		AutoResolve:  autoResolveSvc,
		AI:           aiSvc,
	}
}

type allHandlers struct {
	Auth              *handler.AuthHandler
	User              *handler.UserHandler
	Game              *handler.GameHandler
	Event             *handler.EventHandler
	Leaderboard       *handler.LeaderboardHandler
	Friends           *handler.FriendsHandler
	Reward            *handler.RewardHandler
	Device            *handler.DeviceHandler
	Achievement       *handler.AchievementHandler
	Referral          *handler.ReferralHandler
	League            *handler.LeagueHandler
	WebSocket         *wsHandler.WebSocketHandler
	AdminEvent *adminHandler.EventHandler
	AdminMatch        *adminHandler.MatchHandler
	AdminCard         *adminHandler.CardHandler
	AdminBasket       *adminHandler.BasketHandler
	AdminUser         *adminHandler.UserHandler
	AdminReward       *adminHandler.RewardHandler
	AdminNotification *adminHandler.NotificationHandler
	AdminAudit        *adminHandler.AuditHandler
	AdminDashboard    *adminHandler.DashboardHandler
	AdminAutomation   *adminHandler.AutomationHandler
	AdminSettings     *adminHandler.SettingsHandler
	Exchange          *handler.ExchangeHandler
}

func initHandlers(cfg *config.Config, svc *services, repos *repositories, wsHub *ws.Hub) *allHandlers {
	return &allHandlers{
		Auth:              handler.NewAuthHandler(svc.Auth),
		User:              handler.NewUserHandler(repos.User),
		Game:              handler.NewGameHandler(svc.Game),
		Event:             handler.NewEventHandler(repos.Event),
		Leaderboard:       handler.NewLeaderboardHandler(svc.Leaderboard),
		Friends:           handler.NewFriendsHandler(svc.Leaderboard, repos.Referral, repos.MiniLeague),
		Reward:            handler.NewRewardHandler(svc.Reward, svc.Streak),
		Device:            handler.NewDeviceHandler(repos.FCMToken),
		Achievement:       handler.NewAchievementHandler(svc.Achievement),
		Referral:          handler.NewReferralHandler(svc.Referral),
		League:            handler.NewLeagueHandler(svc.MiniLeague),
		WebSocket:         wsHandler.NewWebSocketHandler(wsHub, cfg.JWTSecret),
		AdminEvent:        adminHandler.NewEventHandler(repos.Event),
		AdminMatch:        adminHandler.NewMatchHandler(repos.Match),
		AdminCard:         adminHandler.NewCardHandler(svc.Card, repos.Card),
		AdminBasket:       adminHandler.NewBasketHandler(repos.Basket),
		AdminUser:         adminHandler.NewUserHandler(repos.User),
		AdminReward:       adminHandler.NewRewardHandler(svc.Reward),
		AdminNotification: adminHandler.NewNotificationHandler(svc.Notification, repos.NotificationHistory),
		AdminAudit:        adminHandler.NewAuditHandler(svc.Audit, svc.Abuse),
		AdminDashboard:    adminHandler.NewDashboardHandler(svc.Leaderboard, svc.Audit, repos.User, repos.Session, repos.Answer, repos.Abuse, repos.Referral, repos.Reward, repos.NotificationHistory, repos.PrizePool),
		AdminAutomation:   adminHandler.NewAutomationHandler(repos.Sport, repos.AutomationLog, svc.SportsData, svc.AutoResolve),
		AdminSettings:     adminHandler.NewSettingsHandler(repos.Setting),
		Exchange:          handler.NewExchangeHandler(svc.Reward, repos.User),
	}
}
