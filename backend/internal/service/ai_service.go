package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"
const defaultModel = "claude-haiku-4-5-20251001"

// AIService provides AI-powered content generation using Claude.
type AIService struct {
	apiKey     string
	httpClient *http.Client
}

// NewAIService creates a new AI service.
func NewAIService(apiKey string) *AIService {
	return &AIService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GeneratedCard represents an AI-generated card question.
type GeneratedCard struct {
	QuestionText       map[string]string `json:"question_text"`
	Tier               string            `json:"tier"`
	HighAnswerIsYes    *bool             `json:"high_answer_is_yes"`
	ResolutionCriteria string            `json:"resolution_criteria"`
}

// MatchContext holds match information for AI prompt.
type MatchContext struct {
	HomeTeam  string
	AwayTeam  string
	SportName string
	League    string
	Kickoff   time.Time
	HomeOdds  float64
	AwayOdds  float64
	DrawOdds  float64
}

type anthropicRequest struct {
	Model     string            `json:"model"`
	MaxTokens int               `json:"max_tokens"`
	Messages  []anthropicMsg    `json:"messages"`
}

type anthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
}

func (s *AIService) callClaude(ctx context.Context, prompt string, maxTokens int) (string, error) {
	reqBody := anthropicRequest{
		Model:     defaultModel,
		MaxTokens: maxTokens,
		Messages: []anthropicMsg{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicAPIURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result anthropicResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("empty response from Claude")
	}

	return result.Content[0].Text, nil
}

// GenerateCardQuestions generates prediction card questions for a match.
func (s *AIService) GenerateCardQuestions(ctx context.Context, match MatchContext, goldCount, silverCount, whiteCount int) ([]GeneratedCard, error) {
	totalCount := goldCount + silverCount + whiteCount
	prompt := fmt.Sprintf(cardGenerationPrompt,
		match.SportName, match.League,
		match.HomeTeam, match.AwayTeam,
		match.Kickoff.Format("2006-01-02 15:04 UTC"),
		match.HomeOdds, match.AwayOdds, match.DrawOdds,
		totalCount, goldCount, silverCount, whiteCount,
	)

	text, err := s.callClaude(ctx, prompt, 4096)
	if err != nil {
		return nil, fmt.Errorf("generate cards: %w", err)
	}

	var cards []GeneratedCard
	if err := json.Unmarshal([]byte(text), &cards); err != nil {
		// Try to extract JSON array from response
		start := bytes.IndexByte([]byte(text), '[')
		end := bytes.LastIndexByte([]byte(text), ']')
		if start >= 0 && end > start {
			if err := json.Unmarshal([]byte(text[start:end+1]), &cards); err != nil {
				log.Error().Str("response", text).Msg("failed to parse AI card response")
				return nil, fmt.Errorf("parse card response: %w", err)
			}
		} else {
			return nil, fmt.Errorf("parse card response: %w", err)
		}
	}

	return cards, nil
}

// AutoResolveAnswer determines if a prediction question answer is yes/no based on match results.
func (s *AIService) AutoResolveAnswer(ctx context.Context, homeTeam, awayTeam string, homeScore, awayScore int, questionText, resolutionCriteria string) (bool, error) {
	prompt := fmt.Sprintf(autoResolvePrompt,
		homeTeam, awayTeam,
		homeTeam, homeScore, awayScore, awayTeam,
		questionText, resolutionCriteria,
	)

	text, err := s.callClaude(ctx, prompt, 256)
	if err != nil {
		return false, fmt.Errorf("auto resolve: %w", err)
	}

	var result struct {
		Answer bool `json:"answer"`
	}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		// Try to extract JSON from response
		start := bytes.IndexByte([]byte(text), '{')
		end := bytes.LastIndexByte([]byte(text), '}')
		if start >= 0 && end > start {
			if err := json.Unmarshal([]byte(text[start:end+1]), &result); err != nil {
				return false, fmt.Errorf("parse resolve response: %w", err)
			}
		} else {
			return false, fmt.Errorf("parse resolve response: %w", err)
		}
	}

	return result.Answer, nil
}
