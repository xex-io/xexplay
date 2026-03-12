package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/external/oddsapi"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

// SportsDataService fetches match data from The Odds API and stores it in the database.
type SportsDataService struct {
	oddsAPI   *oddsapi.Client
	matchRepo *postgres.MatchRepo
	eventRepo *postgres.EventRepo
	cardRepo  *postgres.CardRepo
	basketRepo *postgres.BasketRepo
	sportRepo *postgres.SportRepo
	aiSvc     *AIService
	logRepo   *postgres.AutomationLogRepo
}

// NewSportsDataService creates a new sports data service.
func NewSportsDataService(
	oddsAPI *oddsapi.Client,
	matchRepo *postgres.MatchRepo,
	eventRepo *postgres.EventRepo,
	cardRepo *postgres.CardRepo,
	basketRepo *postgres.BasketRepo,
	sportRepo *postgres.SportRepo,
	aiSvc *AIService,
	logRepo *postgres.AutomationLogRepo,
) *SportsDataService {
	return &SportsDataService{
		oddsAPI:    oddsAPI,
		matchRepo:  matchRepo,
		eventRepo:  eventRepo,
		cardRepo:   cardRepo,
		basketRepo: basketRepo,
		sportRepo:  sportRepo,
		aiSvc:      aiSvc,
		logRepo:    logRepo,
	}
}

// FetchUpcomingMatches fetches upcoming matches for all active sports and upserts them.
func (s *SportsDataService) FetchUpcomingMatches(ctx context.Context) error {
	sports, err := s.sportRepo.FindActive(ctx)
	if err != nil {
		return fmt.Errorf("find active sports: %w", err)
	}

	totalFetched := 0
	for _, sport := range sports {
		matches, err := s.oddsAPI.GetUpcomingMatches(ctx, sport.Key)
		if err != nil {
			log.Error().Err(err).Str("sport", sport.Key).Msg("failed to fetch matches")
			continue
		}

		// Get or create event for this sport
		event, err := s.eventRepo.FindOrCreateAutoEvent(ctx, sport.Key, sport.Title)
		if err != nil {
			log.Error().Err(err).Str("sport", sport.Key).Msg("failed to find/create event")
			continue
		}

		// Filter to next 3 days
		var filtered []oddsapi.OddsMatch
		var teamNames []string
		for _, om := range matches {
			if om.CommenceTime.After(time.Now().Add(72 * time.Hour)) {
				continue
			}
			filtered = append(filtered, om)
			teamNames = append(teamNames, om.HomeTeam, om.AwayTeam)
		}

		// Batch-translate all team names for this sport
		var teamTranslations map[string]map[string]string
		if len(teamNames) > 0 && s.aiSvc != nil {
			teamTranslations, err = s.aiSvc.TranslateTeamNames(ctx, teamNames)
			if err != nil {
				log.Warn().Err(err).Str("sport", sport.Key).Msg("failed to translate team names, continuing without translations")
			}
		}

		for _, om := range filtered {
			m := &domain.Match{
				ID:          uuid.New(),
				EventID:     event.ID,
				HomeTeam:    om.HomeTeam,
				AwayTeam:    om.AwayTeam,
				KickoffTime: om.CommenceTime,
				Status:      domain.MatchStatusUpcoming,
				ExternalID:  om.ID,
				SportKey:    sport.Key,
				Source:      "auto",
			}

			// Attach translated team names if available
			if teamTranslations != nil {
				if t, ok := teamTranslations[om.HomeTeam]; ok {
					m.HomeTeamI18n = t
				}
				if t, ok := teamTranslations[om.AwayTeam]; ok {
					m.AwayTeamI18n = t
				}
			}

			if err := s.matchRepo.UpsertFromExternal(ctx, m); err != nil {
				log.Error().Err(err).
					Str("external_id", om.ID).
					Str("match", om.HomeTeam+" vs "+om.AwayTeam).
					Msg("failed to upsert match")
				continue
			}
			totalFetched++
		}

		log.Info().
			Str("sport", sport.Key).
			Int("fetched", len(matches)).
			Msg("fetched matches from Odds API")
	}

	logDetails, _ := json.Marshal(map[string]interface{}{
		"sports_checked":  len(sports),
		"matches_fetched": totalFetched,
	})
	s.logRepo.Create(ctx, &domain.AutomationLog{
		JobName:        "fetchUpcomingMatches",
		Status:         "success",
		Details:        logDetails,
		ItemsProcessed: totalFetched,
	})

	return nil
}

// GenerateDailyCards generates AI prediction cards for tomorrow's matches.
func (s *SportsDataService) GenerateDailyCards(ctx context.Context) error {
	sports, err := s.sportRepo.FindActive(ctx)
	if err != nil {
		return fmt.Errorf("find active sports: %w", err)
	}

	tomorrow := time.Now().UTC().Add(24 * time.Hour)
	tomorrowStart := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, time.UTC)
	tomorrowEnd := tomorrowStart.Add(24 * time.Hour)

	totalCards := 0
	for _, sport := range sports {
		matches, err := s.matchRepo.FindByDateRange(ctx, tomorrowStart, tomorrowEnd, sport.Key)
		if err != nil {
			log.Error().Err(err).Str("sport", sport.Key).Msg("failed to find matches for card generation")
			continue
		}

		if len(matches) == 0 {
			continue
		}

		// Get or create event
		event, err := s.eventRepo.FindOrCreateAutoEvent(ctx, sport.Key, sport.Title)
		if err != nil {
			log.Error().Err(err).Str("sport", sport.Key).Msg("failed to find/create event")
			continue
		}

		// Calculate questions per match
		goldPerMatch, silverPerMatch, whitePerMatch := distributeCardsAcrossMatches(len(matches))

		var allCardIDs []uuid.UUID
		for _, match := range matches {
			// Skip if cards already exist for this match
			existingCount, err := s.cardRepo.CountByMatchID(ctx, match.ID)
			if err != nil {
				log.Error().Err(err).Str("match_id", match.ID.String()).Msg("failed to count existing cards")
				continue
			}
			if existingCount > 0 {
				continue
			}

			// Fetch odds for the match
			homeOdds, awayOdds, drawOdds := 0.0, 0.0, 0.0
			oddsMatches, err := s.oddsAPI.GetUpcomingMatches(ctx, sport.Key)
			if err == nil {
				for _, om := range oddsMatches {
					if om.ID == match.ExternalID {
						homeOdds, awayOdds, drawOdds = om.H2HOdds()
						break
					}
				}
			}

			// Generate cards via AI
			matchCtx := MatchContext{
				HomeTeam:  match.HomeTeam,
				AwayTeam:  match.AwayTeam,
				SportName: sport.Group,
				League:    sport.Title,
				Kickoff:   match.KickoffTime,
				HomeOdds:  homeOdds,
				AwayOdds:  awayOdds,
				DrawOdds:  drawOdds,
			}

			generated, err := s.aiSvc.GenerateCardQuestions(ctx, matchCtx, goldPerMatch, silverPerMatch, whitePerMatch)
			if err != nil {
				log.Error().Err(err).
					Str("match", match.HomeTeam+" vs "+match.AwayTeam).
					Msg("failed to generate card questions")
				continue
			}

			// Store prompt data for traceability
			promptData, _ := json.Marshal(matchCtx)

			for _, gc := range generated {
				questionJSON, err := json.Marshal(gc.QuestionText)
				if err != nil {
					log.Error().Err(err).Msg("failed to marshal question text")
					continue
				}

				card := &domain.Card{
					ID:                 uuid.New(),
					MatchID:            match.ID,
					QuestionText:       questionJSON,
					Tier:               gc.Tier,
					HighAnswerIsYes:    gc.HighAnswerIsYes,
					IsResolved:         false,
					AvailableDate:      tomorrowStart,
					ExpiresAt:          match.KickoffTime,
					Source:             "ai",
					AIPromptData:       promptData,
					ResolutionCriteria: gc.ResolutionCriteria,
				}

				if err := s.cardRepo.Create(ctx, card); err != nil {
					log.Error().Err(err).Msg("failed to create card")
					continue
				}

				allCardIDs = append(allCardIDs, card.ID)
				totalCards++
			}
		}

		// Create basket for the event + date if we generated enough cards
		if len(allCardIDs) >= domain.TotalCards {
			basket := &domain.DailyBasket{
				ID:          uuid.New(),
				BasketDate:  tomorrowStart,
				EventID:     event.ID,
				IsPublished: false,
			}
			if err := s.basketRepo.Create(ctx, basket); err != nil {
				log.Error().Err(err).Str("event", event.SportKey).Msg("failed to create basket")
				continue
			}

			// Add cards to basket (up to TotalCards)
			cardIDs := allCardIDs
			if len(cardIDs) > domain.TotalCards {
				cardIDs = cardIDs[:domain.TotalCards]
			}
			if err := s.basketRepo.AddCards(ctx, basket.ID, cardIDs); err != nil {
				log.Error().Err(err).Msg("failed to add cards to basket")
			}

			log.Info().
				Str("sport", sport.Key).
				Int("cards", len(cardIDs)).
				Str("basket_id", basket.ID.String()).
				Msg("created auto basket")
		}
	}

	logDetails, _ := json.Marshal(map[string]interface{}{
		"cards_generated": totalCards,
	})
	s.logRepo.Create(ctx, &domain.AutomationLog{
		JobName:        "generateDailyCards",
		Status:         "success",
		Details:        logDetails,
		ItemsProcessed: totalCards,
	})

	return nil
}

// AutoPublishBaskets publishes any unpublished auto-generated baskets for today.
func (s *SportsDataService) AutoPublishBaskets(ctx context.Context) error {
	today := time.Now().UTC()
	todayStart := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)

	baskets, err := s.basketRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("find baskets: %w", err)
	}

	published := 0
	for _, basket := range baskets {
		basketDate := time.Date(basket.BasketDate.Year(), basket.BasketDate.Month(), basket.BasketDate.Day(), 0, 0, 0, 0, time.UTC)
		if basket.IsPublished || !basketDate.Equal(todayStart) {
			continue
		}

		if err := s.basketRepo.Publish(ctx, basket.ID); err != nil {
			log.Warn().Err(err).Str("basket_id", basket.ID.String()).Msg("failed to auto-publish basket (may not have correct tier counts)")
			continue
		}
		published++
	}

	if published > 0 {
		logDetails, _ := json.Marshal(map[string]interface{}{
			"baskets_published": published,
		})
		s.logRepo.Create(ctx, &domain.AutomationLog{
			JobName:        "autoPublishBaskets",
			Status:         "success",
			Details:        logDetails,
			ItemsProcessed: published,
		})
	}

	return nil
}

// SyncSports refreshes the sports list from the Odds API.
func (s *SportsDataService) SyncSports(ctx context.Context) error {
	apiSports, err := s.oddsAPI.GetSports(ctx)
	if err != nil {
		return fmt.Errorf("fetch sports: %w", err)
	}

	synced := 0
	for _, as := range apiSports {
		if !as.Active {
			continue
		}
		sport := &domain.Sport{
			Key:      as.Key,
			Group:    as.Group,
			Title:    as.Title,
			IsActive: false, // New sports default to inactive, admin enables them
		}
		// Only upsert — don't change is_active for existing sports
		existing, _ := s.sportRepo.FindByKey(ctx, as.Key)
		if existing != nil {
			sport.IsActive = existing.IsActive
		}
		if err := s.sportRepo.Upsert(ctx, sport); err != nil {
			log.Error().Err(err).Str("key", as.Key).Msg("failed to upsert sport")
			continue
		}
		synced++
	}

	logDetails, _ := json.Marshal(map[string]interface{}{
		"sports_synced": synced,
	})
	s.logRepo.Create(ctx, &domain.AutomationLog{
		JobName:        "syncSports",
		Status:         "success",
		Details:        logDetails,
		ItemsProcessed: synced,
	})

	return nil
}

// distributeCardsAcrossMatches distributes the target tier counts across available matches.
func distributeCardsAcrossMatches(matchCount int) (gold, silver, white int) {
	if matchCount == 0 {
		return 0, 0, 0
	}
	// Target: 3 gold + 5 silver + 7 white = 15 total per event
	// Distribute evenly, with remainder going to first matches
	gold = max(1, domain.GoldCount/matchCount)
	silver = max(1, domain.SilverCount/matchCount)
	white = max(1, domain.WhiteCount/matchCount)
	return
}
