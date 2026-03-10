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
	authed.Use(middleware.Auth(cfg.JWTSecret))
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

	// Admin routes
	admin := router.Group("/v1/admin")
	admin.Use(middleware.Auth(cfg.JWTSecret))
	admin.Use(middleware.Admin())
	admin.Use(middleware.RateLimiter(rdb, 100, time.Minute))
	admin.Use(middleware.AuditLog(services.Audit))
	{
		// Events
		admin.GET("/events", handlers.AdminEvent.List)
		admin.POST("/events", handlers.AdminEvent.Create)
		admin.PUT("/events/:id", handlers.AdminEvent.Update)

		// Matches
		admin.GET("/matches", handlers.AdminMatch.List)
		admin.POST("/matches", handlers.AdminMatch.Create)
		admin.PUT("/matches/:id", handlers.AdminMatch.Update)

		// Cards
		admin.GET("/cards", handlers.AdminCard.List)
		admin.POST("/cards", handlers.AdminCard.Create)
		admin.PUT("/cards/:id", handlers.AdminCard.Update)
		admin.POST("/cards/:id/resolve", handlers.AdminCard.Resolve)

		// Baskets
		admin.GET("/baskets", handlers.AdminBasket.List)
		admin.POST("/baskets", handlers.AdminBasket.Create)
		admin.PUT("/baskets/:id", handlers.AdminBasket.Update)
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
	User        *postgres.UserRepo
	Event       *postgres.EventRepo
	Match       *postgres.MatchRepo
	Card        *postgres.CardRepo
	Basket      *postgres.BasketRepo
	Session     *postgres.SessionRepo
	Answer      *postgres.AnswerRepo
	Leaderboard *postgres.LeaderboardRepo
	Streak      *postgres.StreakRepo
	Reward      *postgres.RewardRepo
	FCMToken    *postgres.FCMTokenRepo
	Achievement *postgres.AchievementRepo
	Referral    *postgres.ReferralRepo
	MiniLeague  *postgres.MiniLeagueRepo
	Audit       *postgres.AuditRepo
	Abuse       *postgres.AbuseRepo
	Cache       *redis.CacheRepo
	LBCache     *redis.LeaderboardCache
}

func initRepositories(db *postgres.DB, rdb *redis.Client) *repositories {
	return &repositories{
		User:        postgres.NewUserRepo(db),
		Event:       postgres.NewEventRepo(db),
		Match:       postgres.NewMatchRepo(db),
		Card:        postgres.NewCardRepo(db),
		Basket:      postgres.NewBasketRepo(db),
		Session:     postgres.NewSessionRepo(db),
		Answer:      postgres.NewAnswerRepo(db),
		Leaderboard: postgres.NewLeaderboardRepo(db),
		Streak:      postgres.NewStreakRepo(db),
		Reward:      postgres.NewRewardRepo(db),
		FCMToken:    postgres.NewFCMTokenRepo(db),
		Achievement: postgres.NewAchievementRepo(db),
		Referral:    postgres.NewReferralRepo(db),
		MiniLeague:  postgres.NewMiniLeagueRepo(db),
		Audit:       postgres.NewAuditRepo(db),
		Abuse:       postgres.NewAbuseRepo(db),
		Cache:       redis.NewCacheRepo(rdb),
		LBCache:     redis.NewLeaderboardCache(rdb),
	}
}

type services struct {
	Auth         *service.AuthService
	Game         *service.GameService
	Card         *service.CardService
	Shuffle      *service.ShuffleService
	Leaderboard  *service.LeaderboardService
	Streak       *service.StreakService
	Reward       *service.RewardService
	Notification *service.NotificationService
	Cron         *service.CronService
	Achievement  *service.AchievementService
	Referral     *service.ReferralService
	MiniLeague   *service.MiniLeagueService
	Audit        *service.AuditService
	Abuse        *service.AbuseService
}

func initServices(cfg *config.Config, repos *repositories, wsHub *ws.Hub) *services {
	shuffleSvc := service.NewShuffleService()
	leaderboardSvc := service.NewLeaderboardService(repos.Leaderboard, repos.LBCache, repos.User, wsHub)
	streakSvc := service.NewStreakService(repos.Streak)
	rewardSvc := service.NewRewardService(repos.Reward, repos.User, wsHub)

	// Notification service with log sender (swap with FCM sender in production)
	logSender := service.NewLogSender()
	notificationSvc := service.NewNotificationService(repos.FCMToken, logSender)

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

	return &services{
		Auth:         service.NewAuthService(repos.User, referralSvc, cfg.JWTSecret),
		Game:         service.NewGameService(repos.Session, repos.Answer, repos.Basket, repos.Card, repos.User, repos.Cache, shuffleSvc, streakSvc, achievementSvc),
		Card:         service.NewCardService(repos.Card, repos.Answer, leaderboardSvc, wsHub),
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
	AdminEvent        *adminHandler.EventHandler
	AdminMatch        *adminHandler.MatchHandler
	AdminCard         *adminHandler.CardHandler
	AdminBasket       *adminHandler.BasketHandler
	AdminUser         *adminHandler.UserHandler
	AdminReward       *adminHandler.RewardHandler
	AdminNotification *adminHandler.NotificationHandler
	AdminAudit        *adminHandler.AuditHandler
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
		AdminCard:         adminHandler.NewCardHandler(svc.Card),
		AdminBasket:       adminHandler.NewBasketHandler(),
		AdminUser:         adminHandler.NewUserHandler(repos.User),
		AdminReward:       adminHandler.NewRewardHandler(svc.Reward),
		AdminNotification: adminHandler.NewNotificationHandler(svc.Notification),
		AdminAudit:        adminHandler.NewAuditHandler(svc.Audit, svc.Abuse),
		Exchange:          handler.NewExchangeHandler(svc.Reward, repos.User),
	}
}
