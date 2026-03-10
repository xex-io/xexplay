package service

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
	"github.com/xex-exchange/xexplay-api/internal/repository/redis"
)

// CronService runs background jobs on a schedule.
type CronService struct {
	leaderboardRepo *postgres.LeaderboardRepo
	leaderboardSvc  *LeaderboardService
	rewardSvc       *RewardService
	streakRepo      *postgres.StreakRepo
	notificationSvc *NotificationService
	lbCache         *redis.LeaderboardCache
}

func NewCronService(
	leaderboardRepo *postgres.LeaderboardRepo,
	leaderboardSvc *LeaderboardService,
	rewardSvc *RewardService,
	streakRepo *postgres.StreakRepo,
	notificationSvc *NotificationService,
	lbCache *redis.LeaderboardCache,
) *CronService {
	return &CronService{
		leaderboardRepo: leaderboardRepo,
		leaderboardSvc:  leaderboardSvc,
		rewardSvc:       rewardSvc,
		streakRepo:      streakRepo,
		notificationSvc: notificationSvc,
		lbCache:         lbCache,
	}
}

// StartCronJobs launches all background cron jobs in goroutines.
// The provided context should be cancelled to stop all jobs gracefully.
func (s *CronService) StartCronJobs(ctx context.Context) {
	log.Info().Msg("starting cron jobs")

	go s.runAtTime(ctx, 0, 0, "daily-rewards", s.dailyRewardJob)       // midnight UTC
	go s.runAtWeekday(ctx, time.Monday, 0, 5, "weekly-rewards", s.weeklyRewardJob) // Monday 00:05 UTC
	go s.runAtTime(ctx, 20, 0, "streak-at-risk", s.streakAtRiskJob)    // 8 PM UTC
	go s.runAtTime(ctx, 9, 0, "basket-ready", s.basketReadyJob)        // 9 AM UTC
}

// runAtTime runs a job every day at the specified hour and minute (UTC).
func (s *CronService) runAtTime(ctx context.Context, hour, minute int, name string, job func(ctx context.Context)) {
	for {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC)
		if !next.After(now) {
			next = next.Add(24 * time.Hour)
		}

		wait := next.Sub(now)
		log.Info().Str("job", name).Time("next_run", next).Dur("wait", wait).Msg("cron job scheduled")

		select {
		case <-ctx.Done():
			log.Info().Str("job", name).Msg("cron job stopped")
			return
		case <-time.After(wait):
			log.Info().Str("job", name).Msg("cron job starting")
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Error().Str("job", name).Interface("panic", r).Msg("cron job panicked")
					}
				}()
				jobCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
				defer cancel()
				job(jobCtx)
			}()
			log.Info().Str("job", name).Msg("cron job completed")
		}
	}
}

// runAtWeekday runs a job on a specific weekday at the specified hour and minute (UTC).
func (s *CronService) runAtWeekday(ctx context.Context, weekday time.Weekday, hour, minute int, name string, job func(ctx context.Context)) {
	for {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC)

		// Advance to the target weekday
		daysUntil := int(weekday - next.Weekday())
		if daysUntil < 0 {
			daysUntil += 7
		}
		next = next.AddDate(0, 0, daysUntil)

		if !next.After(now) {
			next = next.AddDate(0, 0, 7)
		}

		wait := next.Sub(now)
		log.Info().Str("job", name).Time("next_run", next).Dur("wait", wait).Msg("cron job scheduled")

		select {
		case <-ctx.Done():
			log.Info().Str("job", name).Msg("cron job stopped")
			return
		case <-time.After(wait):
			log.Info().Str("job", name).Msg("cron job starting")
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Error().Str("job", name).Interface("panic", r).Msg("cron job panicked")
					}
				}()
				jobCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
				defer cancel()
				job(jobCtx)
			}()
			log.Info().Str("job", name).Msg("cron job completed")
		}
	}
}

// dailyRewardJob distributes rewards for yesterday's daily leaderboard and resets the Redis key.
func (s *CronService) dailyRewardJob(ctx context.Context) {
	yesterday := time.Now().UTC().AddDate(0, 0, -1)
	periodKey := GetDailyKey(yesterday)

	log.Info().Str("period_key", periodKey).Msg("distributing daily rewards")

	// Get top entries from leaderboard
	entries, err := s.leaderboardRepo.GetTopN(ctx, domain.PeriodDaily, periodKey, 100)
	if err != nil {
		log.Error().Err(err).Str("period_key", periodKey).Msg("failed to get daily leaderboard for reward distribution")
		return
	}

	if len(entries) == 0 {
		log.Info().Str("period_key", periodKey).Msg("no daily leaderboard entries, skipping reward distribution")
		return
	}

	// Convert to reward entries
	rewardEntries := make([]domain.RewardLeaderboardEntry, 0, len(entries))
	for _, e := range entries {
		rewardEntries = append(rewardEntries, domain.RewardLeaderboardEntry{
			UserID:      e.UserID,
			Rank:        e.Rank,
			TotalPoints: e.TotalPoints,
		})
	}

	count, err := s.rewardSvc.DistributeRewards(ctx, domain.PeriodDaily, periodKey, rewardEntries)
	if err != nil {
		log.Error().Err(err).Str("period_key", periodKey).Msg("failed to distribute daily rewards")
		return
	}

	log.Info().Int("distributed", count).Str("period_key", periodKey).Msg("daily rewards distributed")

	// Notify rewarded users
	for _, entry := range rewardEntries {
		s.notificationSvc.NotifyRewardEarned(ctx, entry.UserID, "daily_leaderboard", float64(entry.Rank))
	}

	// Clean up yesterday's Redis leaderboard key
	key := redis.LeaderboardKey(domain.PeriodDaily, periodKey)
	if err := s.lbCache.DeleteKey(ctx, key); err != nil {
		log.Warn().Err(err).Str("key", key).Msg("failed to delete daily leaderboard cache key")
	}
}

// weeklyRewardJob distributes rewards for last week's leaderboard and resets the Redis key.
func (s *CronService) weeklyRewardJob(ctx context.Context) {
	lastWeek := time.Now().UTC().AddDate(0, 0, -7)
	periodKey := GetWeeklyKey(lastWeek)

	log.Info().Str("period_key", periodKey).Msg("distributing weekly rewards")

	entries, err := s.leaderboardRepo.GetTopN(ctx, domain.PeriodWeekly, periodKey, 100)
	if err != nil {
		log.Error().Err(err).Str("period_key", periodKey).Msg("failed to get weekly leaderboard for reward distribution")
		return
	}

	if len(entries) == 0 {
		log.Info().Str("period_key", periodKey).Msg("no weekly leaderboard entries, skipping reward distribution")
		return
	}

	rewardEntries := make([]domain.RewardLeaderboardEntry, 0, len(entries))
	for _, e := range entries {
		rewardEntries = append(rewardEntries, domain.RewardLeaderboardEntry{
			UserID:      e.UserID,
			Rank:        e.Rank,
			TotalPoints: e.TotalPoints,
		})
	}

	count, err := s.rewardSvc.DistributeRewards(ctx, domain.PeriodWeekly, periodKey, rewardEntries)
	if err != nil {
		log.Error().Err(err).Str("period_key", periodKey).Msg("failed to distribute weekly rewards")
		return
	}

	log.Info().Int("distributed", count).Str("period_key", periodKey).Msg("weekly rewards distributed")

	// Notify rewarded users
	for _, entry := range rewardEntries {
		s.notificationSvc.NotifyRewardEarned(ctx, entry.UserID, "weekly_leaderboard", float64(entry.Rank))
	}

	// Clean up last week's Redis leaderboard key
	key := redis.LeaderboardKey(domain.PeriodWeekly, periodKey)
	if err := s.lbCache.DeleteKey(ctx, key); err != nil {
		log.Warn().Err(err).Str("key", key).Msg("failed to delete weekly leaderboard cache key")
	}
}

// streakAtRiskJob finds users who played yesterday but not today and notifies them.
func (s *CronService) streakAtRiskJob(ctx context.Context) {
	log.Info().Msg("checking for streaks at risk")

	streaks, err := s.streakRepo.FindStreaksAtRisk(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to find streaks at risk")
		return
	}

	log.Info().Int("count", len(streaks)).Msg("found streaks at risk")

	for _, streak := range streaks {
		s.notificationSvc.NotifyStreakAtRisk(ctx, streak.UserID, streak.CurrentStreak)
	}
}

// basketReadyJob sends a broadcast notification that the daily basket is available.
func (s *CronService) basketReadyJob(ctx context.Context) {
	log.Info().Msg("sending basket ready notification")
	s.notificationSvc.NotifyBasketReady(ctx)
}
