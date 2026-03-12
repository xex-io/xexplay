package oddsapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

const baseURL = "https://api.the-odds-api.com"

// Client is an HTTP client for The Odds API.
type Client struct {
	apiKey     string
	httpClient *http.Client
	remaining  int // remaining API requests this month
}

// NewClient creates a new Odds API client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		remaining: -1,
	}
}

// RemainingRequests returns the number of remaining API requests this month, or -1 if unknown.
func (c *Client) RemainingRequests() int {
	return c.remaining
}

func (c *Client) doRequest(ctx context.Context, path string, params map[string]string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	q := req.URL.Query()
	q.Set("apiKey", c.apiKey)
	for k, v := range params {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	// Track remaining requests
	if rem := resp.Header.Get("x-requests-remaining"); rem != "" {
		if n, err := strconv.Atoi(rem); err == nil {
			c.remaining = n
			if n < 500 {
				log.Warn().Int("remaining", n).Msg("Odds API requests running low")
			}
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// GetSports returns all available sports.
func (c *Client) GetSports(ctx context.Context) ([]SportResponse, error) {
	body, err := c.doRequest(ctx, "/v4/sports", nil)
	if err != nil {
		return nil, err
	}

	var sports []SportResponse
	if err := json.Unmarshal(body, &sports); err != nil {
		return nil, fmt.Errorf("decode sports: %w", err)
	}
	return sports, nil
}

// GetUpcomingMatches returns upcoming matches with odds for a given sport.
func (c *Client) GetUpcomingMatches(ctx context.Context, sportKey string) ([]OddsMatch, error) {
	params := map[string]string{
		"regions":  "eu,us",
		"markets":  "h2h",
		"oddsFormat": "decimal",
	}

	body, err := c.doRequest(ctx, "/v4/sports/"+sportKey+"/odds", params)
	if err != nil {
		return nil, err
	}

	var matches []OddsMatch
	if err := json.Unmarshal(body, &matches); err != nil {
		return nil, fmt.Errorf("decode matches: %w", err)
	}
	return matches, nil
}

// GetScores returns scores for completed and in-progress matches.
func (c *Client) GetScores(ctx context.Context, sportKey string, daysFrom int) ([]ScoreResult, error) {
	params := map[string]string{
		"daysFrom": strconv.Itoa(daysFrom),
	}

	body, err := c.doRequest(ctx, "/v4/sports/"+sportKey+"/scores", params)
	if err != nil {
		return nil, err
	}

	var scores []ScoreResult
	if err := json.Unmarshal(body, &scores); err != nil {
		return nil, fmt.Errorf("decode scores: %w", err)
	}
	return scores, nil
}
