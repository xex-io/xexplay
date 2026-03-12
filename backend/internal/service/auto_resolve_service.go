package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/external/oddsapi"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

// AutoResolveService handles automatic resolution of cards based on match results.
type AutoResolveService struct {
	matchRepo *postgres.MatchRepo
	cardRepo  *postgres.CardRepo
	cardSvc   *CardService
	oddsAPI   *oddsapi.Client
	aiSvc     *AIService
	logRepo   *postgres.AutomationLogRepo
}

// NewAutoResolveService creates a new auto-resolve service.
func NewAutoResolveService(
	matchRepo *postgres.MatchRepo,
	cardRepo *postgres.CardRepo,
	cardSvc *CardService,
	oddsAPI *oddsapi.Client,
	aiSvc *AIService,
	logRepo *postgres.AutomationLogRepo,
) *AutoResolveService {
	return &AutoResolveService{
		matchRepo: matchRepo,
		cardRepo:  cardRepo,
		cardSvc:   cardSvc,
		oddsAPI:   oddsAPI,
		aiSvc:     aiSvc,
		logRepo:   logRepo,
	}
}

// ProcessCompletedMatches fetches scores for pending matches and auto-resolves their cards.
func (s *AutoResolveService) ProcessCompletedMatches(ctx context.Context) error {
	// Find matches that are past kickoff but not completed
	matches, err := s.matchRepo.FindScheduledPastKickoff(ctx)
	if err != nil {
		return fmt.Errorf("find pending matches: %w", err)
	}

	if len(matches) == 0 {
		return nil
	}

	// Group by sport_key to minimize API calls
	bySport := make(map[string][]*domain.Match)
	for _, m := range matches {
		if m.SportKey != "" {
			bySport[m.SportKey] = append(bySport[m.SportKey], m)
		}
	}

	resolved := 0
	for sportKey, sportMatches := range bySport {
		scores, err := s.oddsAPI.GetScores(ctx, sportKey, 3)
		if err != nil {
			log.Error().Err(err).Str("sport", sportKey).Msg("failed to fetch scores")
			continue
		}

		// Index scores by external ID
		scoreMap := make(map[string]*oddsapi.ScoreResult)
		for i := range scores {
			scoreMap[scores[i].ID] = &scores[i]
		}

		for _, match := range sportMatches {
			if match.ExternalID == "" {
				continue
			}

			score, ok := scoreMap[match.ExternalID]
			if !ok || !score.Completed {
				// Update to live if we have scores but not completed
				if ok && !score.Completed && match.Status != domain.MatchStatusLive {
					match.Status = domain.MatchStatusLive
					homeScore, awayScore := parseScores(score, match.HomeTeam, match.AwayTeam)
					match.HomeScore = &homeScore
					match.AwayScore = &awayScore
					if err := s.matchRepo.Update(ctx, match); err != nil {
						log.Error().Err(err).Str("match_id", match.ID.String()).Msg("failed to update live match")
					}
				}
				continue
			}

			// Match is completed — update score
			homeScore, awayScore := parseScores(score, match.HomeTeam, match.AwayTeam)
			if err := s.matchRepo.UpdateResult(ctx, match.ID, homeScore, awayScore); err != nil {
				log.Error().Err(err).Str("match_id", match.ID.String()).Msg("failed to update match result")
				continue
			}

			// Auto-resolve cards
			cards, err := s.cardRepo.FindUnresolvedByMatch(ctx, match.ID)
			if err != nil {
				log.Error().Err(err).Str("match_id", match.ID.String()).Msg("failed to find unresolved cards")
				continue
			}

			for _, card := range cards {
				answer, err := s.resolveCard(ctx, card, match, homeScore, awayScore)
				if err != nil {
					log.Error().Err(err).
						Str("card_id", card.ID.String()).
						Str("match_id", match.ID.String()).
						Msg("failed to auto-resolve card")
					continue
				}

				if err := s.cardSvc.ResolveCard(ctx, card.ID, answer); err != nil {
					log.Error().Err(err).
						Str("card_id", card.ID.String()).
						Bool("answer", answer).
						Msg("failed to resolve card via CardService")
					continue
				}

				resolved++
				log.Info().
					Str("card_id", card.ID.String()).
					Bool("answer", answer).
					Str("match", match.HomeTeam+" vs "+match.AwayTeam).
					Msg("auto-resolved card")
			}
		}
	}

	// Log the result
	logDetails, _ := json.Marshal(map[string]interface{}{
		"matches_checked": len(matches),
		"cards_resolved":  resolved,
	})
	s.logRepo.Create(ctx, &domain.AutomationLog{
		JobName:        "autoResolveCards",
		Status:         "success",
		Details:        logDetails,
		ItemsProcessed: resolved,
	})

	return nil
}

// resolveCard determines the correct answer for a card based on match results.
func (s *AutoResolveService) resolveCard(ctx context.Context, card *domain.Card, match *domain.Match, homeScore, awayScore int) (bool, error) {
	// Extract English question text
	var questionTexts map[string]string
	if err := json.Unmarshal(card.QuestionText, &questionTexts); err != nil {
		return false, fmt.Errorf("parse question text: %w", err)
	}

	questionEn := questionTexts["en"]
	if questionEn == "" {
		return false, fmt.Errorf("no English question text found")
	}

	// Use AI to determine the answer
	answer, err := s.aiSvc.AutoResolveAnswer(
		ctx,
		match.HomeTeam, match.AwayTeam,
		homeScore, awayScore,
		questionEn, card.ResolutionCriteria,
	)
	if err != nil {
		return false, fmt.Errorf("AI auto-resolve: %w", err)
	}

	return answer, nil
}

// parseScores extracts home and away scores from an Odds API score result.
func parseScores(score *oddsapi.ScoreResult, homeTeam, awayTeam string) (int, int) {
	var homeScore, awayScore int
	for _, ts := range score.Scores {
		s, _ := strconv.Atoi(ts.Score)
		switch ts.Name {
		case homeTeam:
			homeScore = s
		case awayTeam:
			awayScore = s
		}
	}
	return homeScore, awayScore
}
