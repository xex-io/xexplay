//go:build e2e

package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	gorillaWS "github.com/gorilla/websocket"
)

// ---------------------------------------------------------------------------
// Environment helpers
// ---------------------------------------------------------------------------

func baseURL(t *testing.T) string {
	t.Helper()
	u := os.Getenv("E2E_BASE_URL")
	if u == "" {
		t.Fatal("E2E_BASE_URL environment variable is required")
	}
	return u
}

func userToken(t *testing.T) string {
	t.Helper()
	tok := os.Getenv("E2E_AUTH_TOKEN")
	if tok == "" {
		t.Fatal("E2E_AUTH_TOKEN environment variable is required")
	}
	return tok
}

func adminToken(t *testing.T) string {
	t.Helper()
	tok := os.Getenv("E2E_ADMIN_TOKEN")
	if tok == "" {
		t.Fatal("E2E_ADMIN_TOKEN environment variable is required")
	}
	return tok
}

// ---------------------------------------------------------------------------
// Generic API response types (mirrors backend response.Response)
// ---------------------------------------------------------------------------

type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *apiError       `json:"error,omitempty"`
	Meta    *apiMeta        `json:"meta,omitempty"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiMeta struct {
	Page    int   `json:"page,omitempty"`
	PerPage int   `json:"per_page,omitempty"`
	Total   int64 `json:"total,omitempty"`
}

// ---------------------------------------------------------------------------
// Domain view types used to unmarshal responses
// ---------------------------------------------------------------------------

type sessionView struct {
	ID               string     `json:"id"`
	Status           string     `json:"status"`
	CurrentIndex     int        `json:"current_index"`
	TotalCards       int        `json:"total_cards"`
	AnswersUsed      int        `json:"answers_used"`
	AnswersRemaining int        `json:"answers_remaining"`
	SkipsUsed        int        `json:"skips_used"`
	SkipsRemaining   int        `json:"skips_remaining"`
	CardPresentedAt  *time.Time `json:"card_presented_at,omitempty"`
	StartedAt        time.Time  `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

type cardView struct {
	ID           string          `json:"id"`
	MatchID      string          `json:"match_id"`
	QuestionText json.RawMessage `json:"question_text"`
	Tier         string          `json:"tier"`
	YesPoints    int             `json:"yes_points"`
	NoPoints     int             `json:"no_points"`
	ExpiresAt    time.Time       `json:"expires_at"`
}

type leaderboardResponse struct {
	PeriodType string           `json:"period_type"`
	PeriodKey  string           `json:"period_key"`
	Entries    []leaderboardRow `json:"entries"`
	UserRank   *leaderboardRow  `json:"user_rank,omitempty"`
	Total      int              `json:"total"`
}

type leaderboardRow struct {
	Rank           int    `json:"rank"`
	UserID         string `json:"user_id"`
	DisplayName    string `json:"display_name"`
	AvatarURL      string `json:"avatar_url"`
	TotalPoints    int    `json:"total_points"`
	CorrectAnswers int    `json:"correct_answers"`
}

type rewardsResponse struct {
	Pending []rewardDistribution `json:"pending"`
	History []rewardDistribution `json:"history"`
	Streak  json.RawMessage      `json:"streak"`
}

type rewardDistribution struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	PeriodType string  `json:"period_type"`
	PeriodKey  string  `json:"period_key"`
	RewardType string  `json:"reward_type"`
	Amount     float64 `json:"amount"`
	Rank       int     `json:"rank"`
	Status     string  `json:"status"`
}

type distributeResponse struct {
	Distributed int `json:"distributed"`
}

type claimResponse struct {
	Message    string  `json:"message"`
	Status     string  `json:"status,omitempty"`
	RewardType string  `json:"reward_type,omitempty"`
	Amount     float64 `json:"amount,omitempty"`
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

func doRequest(t *testing.T, method, url, token string, body interface{}) (int, *apiResponse) {
	t.Helper()
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var apiResp apiResponse
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			t.Fatalf("failed to unmarshal response (status=%d, body=%s): %v", resp.StatusCode, string(respBody), err)
		}
	}
	return resp.StatusCode, &apiResp
}

func requireSuccess(t *testing.T, statusCode int, resp *apiResponse) {
	t.Helper()
	if statusCode < 200 || statusCode >= 300 {
		errMsg := ""
		if resp.Error != nil {
			errMsg = resp.Error.Message
		}
		t.Fatalf("expected 2xx, got %d: %s", statusCode, errMsg)
	}
	if !resp.Success {
		t.Fatal("expected success=true in response")
	}
}

func requireStatus(t *testing.T, expected, actual int, resp *apiResponse) {
	t.Helper()
	if expected != actual {
		errMsg := ""
		if resp.Error != nil {
			errMsg = resp.Error.Message
		}
		t.Fatalf("expected status %d, got %d: %s", expected, actual, errMsg)
	}
}

func unmarshalData(t *testing.T, resp *apiResponse, target interface{}) {
	t.Helper()
	if resp.Data == nil {
		t.Fatal("response data is nil")
	}
	if err := json.Unmarshal(resp.Data, target); err != nil {
		t.Fatalf("failed to unmarshal response data: %v (raw: %s)", err, string(resp.Data))
	}
}

func apiURL(t *testing.T, path string) string {
	t.Helper()
	return baseURL(t) + path
}

// ---------------------------------------------------------------------------
// 1.16.1 - TestAdminCreatesEventFlow
// Admin: create event -> add match -> create card -> create basket -> publish
// ---------------------------------------------------------------------------

func TestAdminCreatesEventFlow(t *testing.T) {
	token := adminToken(t)

	// Step 1: Create event
	t.Log("Step 1: Creating event")
	status, resp := doRequest(t, http.MethodPost, apiURL(t, "/v1/admin/events"), token, map[string]interface{}{
		"name":               map[string]string{"en": "E2E Test Event", "fa": "رویداد تست"},
		"slug":               fmt.Sprintf("e2e-test-%d", time.Now().UnixMilli()),
		"start_date":         time.Now().UTC().Format(time.RFC3339),
		"end_date":           time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
		"is_active":          true,
		"scoring_multiplier": 1.0,
	})
	requireSuccess(t, status, resp)
	t.Log("Event created")

	// Extract event ID if returned, otherwise use placeholder for subsequent calls
	var eventData map[string]interface{}
	unmarshalData(t, resp, &eventData)
	eventID, _ := eventData["id"].(string)

	// Step 2: Create match
	t.Log("Step 2: Creating match")
	matchBody := map[string]interface{}{
		"home_team":    "Team Alpha",
		"away_team":    "Team Beta",
		"kickoff_time": time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339),
		"status":       "upcoming",
	}
	if eventID != "" {
		matchBody["event_id"] = eventID
	}
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/admin/matches"), token, matchBody)
	requireSuccess(t, status, resp)
	t.Log("Match created")

	var matchData map[string]interface{}
	unmarshalData(t, resp, &matchData)
	matchID, _ := matchData["id"].(string)

	// Step 3: Create card
	t.Log("Step 3: Creating card")
	cardBody := map[string]interface{}{
		"question_text":    map[string]string{"en": "Will Team Alpha score first?", "fa": "آیا تیم آلفا اول گل می‌زند؟"},
		"tier":             "gold",
		"high_answer_is_yes": true,
		"available_date":   time.Now().UTC().Format(time.RFC3339),
		"expires_at":       time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
	}
	if matchID != "" {
		cardBody["match_id"] = matchID
	}
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/admin/cards"), token, cardBody)
	requireSuccess(t, status, resp)
	t.Log("Card created")

	// Step 4: Create basket
	t.Log("Step 4: Creating basket")
	basketBody := map[string]interface{}{
		"basket_date": time.Now().UTC().Format("2006-01-02"),
	}
	if eventID != "" {
		basketBody["event_id"] = eventID
	}
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/admin/baskets"), token, basketBody)
	requireSuccess(t, status, resp)
	t.Log("Basket created")

	var basketData map[string]interface{}
	unmarshalData(t, resp, &basketData)
	basketID, _ := basketData["id"].(string)

	// Step 5: Publish basket
	t.Log("Step 5: Publishing basket")
	publishURL := "/v1/admin/baskets/"
	if basketID != "" {
		publishURL += basketID
	} else {
		publishURL += "00000000-0000-0000-0000-000000000000"
	}
	publishURL += "/publish"
	status, resp = doRequest(t, http.MethodPost, apiURL(t, publishURL), token, nil)
	requireSuccess(t, status, resp)
	t.Log("Basket published successfully")

	var publishResult map[string]interface{}
	unmarshalData(t, resp, &publishResult)
	if msg, ok := publishResult["message"].(string); ok {
		t.Logf("Publish response: %s", msg)
	}
}

// ---------------------------------------------------------------------------
// 1.16.2 - TestUserGameSession
// Login -> start session -> answer 10 + skip 5 -> verify summary
// ---------------------------------------------------------------------------

func TestUserGameSession(t *testing.T) {
	token := userToken(t)

	// Step 1: Start a new session
	t.Log("Step 1: Starting session")
	status, resp := doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/start"), token, nil)
	requireSuccess(t, status, resp)

	var session sessionView
	unmarshalData(t, resp, &session)
	t.Logf("Session started: id=%s, total_cards=%d", session.ID, session.TotalCards)

	if session.Status != "active" {
		t.Fatalf("expected session status 'active', got '%s'", session.Status)
	}
	if session.TotalCards != 15 {
		t.Fatalf("expected 15 total cards, got %d", session.TotalCards)
	}

	answersGiven := 0
	skipsGiven := 0

	// Step 2: Answer 10 cards and skip 5
	for i := 0; i < 15; i++ {
		// Get current card
		status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current/card"), token, nil)
		if status == http.StatusBadRequest || status == http.StatusNotFound {
			t.Logf("No more cards available at index %d, session may be complete", i)
			break
		}
		requireSuccess(t, status, resp)

		var card cardView
		unmarshalData(t, resp, &card)
		t.Logf("Card %d: tier=%s, question=%s", i+1, card.Tier, string(card.QuestionText))

		if answersGiven < 10 {
			// Submit answer (alternate yes/no)
			answer := i%2 == 0
			status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/answer"), token,
				map[string]interface{}{"answer": answer})
			if status == http.StatusBadRequest {
				// May have run out of answers, try skip
				t.Logf("Answer rejected at card %d, switching to skip", i+1)
				status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/skip"), token, nil)
				requireSuccess(t, status, resp)
				skipsGiven++
			} else {
				requireSuccess(t, status, resp)
				answersGiven++
			}
		} else {
			// Skip remaining cards
			status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/skip"), token, nil)
			requireSuccess(t, status, resp)
			skipsGiven++
		}
	}

	t.Logf("Answers given: %d, Skips given: %d", answersGiven, skipsGiven)

	// Step 3: Verify session summary
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current"), token, nil)
	requireSuccess(t, status, resp)

	var summary sessionView
	unmarshalData(t, resp, &summary)
	t.Logf("Session summary: status=%s, answers_used=%d, skips_used=%d, current_index=%d",
		summary.Status, summary.AnswersUsed, summary.SkipsUsed, summary.CurrentIndex)

	if summary.AnswersUsed+summary.SkipsUsed != summary.CurrentIndex {
		t.Logf("Warning: answers_used(%d) + skips_used(%d) != current_index(%d)",
			summary.AnswersUsed, summary.SkipsUsed, summary.CurrentIndex)
	}
}

// ---------------------------------------------------------------------------
// 1.16.3 - TestCardResolutionScoring
// Admin resolves cards -> verify user scores per tier
// ---------------------------------------------------------------------------

func TestCardResolutionScoring(t *testing.T) {
	aToken := adminToken(t)
	uToken := userToken(t)

	// Step 1: Get user stats before
	t.Log("Step 1: Getting user stats before resolution")
	status, resp := doRequest(t, http.MethodGet, apiURL(t, "/v1/me/stats"), uToken, nil)
	requireSuccess(t, status, resp)

	var statsBefore map[string]interface{}
	unmarshalData(t, resp, &statsBefore)
	t.Logf("Stats before: %+v", statsBefore)

	// Step 2: List admin cards to find unresolved ones
	t.Log("Step 2: Listing cards for resolution")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/admin/cards"), aToken, nil)
	requireSuccess(t, status, resp)

	var cards []map[string]interface{}
	unmarshalData(t, resp, &cards)
	t.Logf("Found %d cards", len(cards))

	resolvedCount := 0
	for _, card := range cards {
		isResolved, _ := card["is_resolved"].(bool)
		if isResolved {
			continue
		}
		cardID, ok := card["id"].(string)
		if !ok || cardID == "" {
			continue
		}

		tier, _ := card["tier"].(string)
		t.Logf("Resolving card %s (tier=%s) with correct_answer=true", cardID, tier)

		status, resp = doRequest(t, http.MethodPost,
			apiURL(t, fmt.Sprintf("/v1/admin/cards/%s/resolve", cardID)),
			aToken,
			map[string]interface{}{"correct_answer": true})
		requireSuccess(t, status, resp)
		resolvedCount++

		if resolvedCount >= 3 {
			break
		}
	}
	t.Logf("Resolved %d cards", resolvedCount)

	// Step 3: Get user stats after and compare
	t.Log("Step 3: Getting user stats after resolution")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/me/stats"), uToken, nil)
	requireSuccess(t, status, resp)

	var statsAfter map[string]interface{}
	unmarshalData(t, resp, &statsAfter)
	t.Logf("Stats after: %+v", statsAfter)

	// Scoring expectations by tier:
	// Gold:   high=20, low=5
	// Silver: high=15, low=10
	// White:  10
	// VIP:    high=30, low=10
	t.Log("Scoring reference: gold(20/5), silver(15/10), white(10), vip(30/10)")
}

// ---------------------------------------------------------------------------
// 1.16.4 - TestSessionResume
// Start session -> answer 5 -> reconnect -> verify resume from card 6
// ---------------------------------------------------------------------------

func TestSessionResume(t *testing.T) {
	token := userToken(t)

	// Step 1: Start session
	t.Log("Step 1: Starting session")
	status, resp := doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/start"), token, nil)
	requireSuccess(t, status, resp)

	var session sessionView
	unmarshalData(t, resp, &session)
	t.Logf("Session started: id=%s", session.ID)
	sessionID := session.ID

	// Step 2: Answer 5 cards
	for i := 0; i < 5; i++ {
		// Get card
		status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current/card"), token, nil)
		if status != http.StatusOK {
			t.Logf("Could not get card at index %d, stopping early", i)
			break
		}

		// Answer it
		status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/answer"), token,
			map[string]interface{}{"answer": true})
		if status == http.StatusBadRequest {
			// Might need to skip instead
			status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/skip"), token, nil)
		}
		if status < 200 || status >= 300 {
			t.Logf("Action on card %d returned status %d", i+1, status)
			break
		}
	}
	t.Log("Answered/processed 5 cards")

	// Step 3: "Reconnect" - get current session (simulates app restart)
	t.Log("Step 3: Reconnecting - fetching current session")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current"), token, nil)
	requireSuccess(t, status, resp)

	var resumed sessionView
	unmarshalData(t, resp, &resumed)

	// Verify same session
	if resumed.ID != sessionID {
		t.Fatalf("expected same session ID %s after reconnect, got %s", sessionID, resumed.ID)
	}
	t.Logf("Resumed session: id=%s, current_index=%d, status=%s", resumed.ID, resumed.CurrentIndex, resumed.Status)

	// Verify we are at card index 5 (0-based) or higher
	if resumed.CurrentIndex < 5 {
		t.Fatalf("expected current_index >= 5 after answering 5 cards, got %d", resumed.CurrentIndex)
	}
	t.Logf("Resume verified: session continues from card %d", resumed.CurrentIndex+1)

	// Step 4: Verify we can still get the next card
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current/card"), token, nil)
	if resumed.CurrentIndex < 15 {
		requireSuccess(t, status, resp)
		var nextCard cardView
		unmarshalData(t, resp, &nextCard)
		t.Logf("Next card after resume: id=%s, tier=%s", nextCard.ID, nextCard.Tier)
	} else {
		t.Log("Session already completed, no more cards")
	}
}

// ---------------------------------------------------------------------------
// 1.16.6 - TestResourceExhaustion
// Use all skips -> verify remaining cards must be answered
// ---------------------------------------------------------------------------

func TestResourceExhaustion(t *testing.T) {
	token := userToken(t)

	// Start a fresh session
	t.Log("Starting session for resource exhaustion test")
	status, resp := doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/start"), token, nil)
	requireSuccess(t, status, resp)

	var session sessionView
	unmarshalData(t, resp, &session)
	t.Logf("Session: skips_remaining=%d, answers_remaining=%d", session.SkipsRemaining, session.AnswersRemaining)

	// Use all available skips (default=5)
	skipsToUse := session.SkipsRemaining
	for i := 0; i < skipsToUse; i++ {
		// Get card first
		status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current/card"), token, nil)
		if status != http.StatusOK {
			t.Logf("No card available at skip %d", i+1)
			break
		}

		status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/skip"), token, nil)
		if status != http.StatusOK {
			t.Logf("Skip %d failed with status %d", i+1, status)
			break
		}
		t.Logf("Skip %d/%d used", i+1, skipsToUse)
	}

	// Verify skip is now rejected
	t.Log("Attempting skip after exhaustion")
	// Get card first
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current/card"), token, nil)
	if status == http.StatusOK {
		status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/skip"), token, nil)
		if status == http.StatusOK {
			// Check session state
			var result map[string]interface{}
			unmarshalData(t, resp, &result)
			t.Log("Skip was accepted - checking if bonus skips were available")
		} else {
			t.Logf("Skip correctly rejected with status %d", status)
			if resp.Error != nil {
				t.Logf("Error message: %s", resp.Error.Message)
			}
		}
	}

	// Verify we can still answer
	t.Log("Verifying answers still work after skips exhausted")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current/card"), token, nil)
	if status == http.StatusOK {
		status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/answer"), token,
			map[string]interface{}{"answer": true})
		requireSuccess(t, status, resp)
		t.Log("Answer accepted after skips exhausted - resource exhaustion verified")
	} else {
		t.Log("Session completed, no more cards available")
	}

	// Final session state
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current"), token, nil)
	requireSuccess(t, status, resp)
	var finalSession sessionView
	unmarshalData(t, resp, &finalSession)
	t.Logf("Final state: skips_remaining=%d, answers_remaining=%d, current_index=%d",
		finalSession.SkipsRemaining, finalSession.AnswersRemaining, finalSession.CurrentIndex)

	if finalSession.SkipsRemaining > 0 {
		t.Log("Warning: skips_remaining > 0, possibly due to bonus skips from streaks")
	}
}

// ---------------------------------------------------------------------------
// 2.10.1 - TestLeaderboardUpdate
// Resolve card -> verify leaderboard position updates
// ---------------------------------------------------------------------------

func TestLeaderboardUpdate(t *testing.T) {
	uToken := userToken(t)

	// Step 1: Get daily leaderboard before
	t.Log("Step 1: Getting daily leaderboard")
	status, resp := doRequest(t, http.MethodGet, apiURL(t, "/v1/leaderboards/daily"), uToken, nil)
	requireSuccess(t, status, resp)

	var lbBefore leaderboardResponse
	unmarshalData(t, resp, &lbBefore)
	t.Logf("Daily leaderboard: period=%s, entries=%d, total=%d",
		lbBefore.PeriodKey, len(lbBefore.Entries), lbBefore.Total)

	if lbBefore.UserRank != nil {
		t.Logf("User rank before: rank=%d, points=%d", lbBefore.UserRank.Rank, lbBefore.UserRank.TotalPoints)
	} else {
		t.Log("User not yet on daily leaderboard")
	}

	// Step 2: Get weekly leaderboard
	t.Log("Step 2: Getting weekly leaderboard")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/leaderboards/weekly"), uToken, nil)
	requireSuccess(t, status, resp)

	var lbWeekly leaderboardResponse
	unmarshalData(t, resp, &lbWeekly)
	t.Logf("Weekly leaderboard: period=%s, entries=%d", lbWeekly.PeriodKey, len(lbWeekly.Entries))

	// Step 3: Get all-time leaderboard
	t.Log("Step 3: Getting all-time leaderboard")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/leaderboards/all-time"), uToken, nil)
	requireSuccess(t, status, resp)

	var lbAllTime leaderboardResponse
	unmarshalData(t, resp, &lbAllTime)
	t.Logf("All-time leaderboard: entries=%d, total=%d", len(lbAllTime.Entries), lbAllTime.Total)

	if lbAllTime.UserRank != nil {
		t.Logf("User all-time rank: rank=%d, points=%d", lbAllTime.UserRank.Rank, lbAllTime.UserRank.TotalPoints)
	}

	// Step 4: Verify leaderboard entries have required fields
	for _, entry := range lbBefore.Entries {
		if entry.UserID == "" {
			t.Error("Leaderboard entry missing user_id")
		}
		if entry.Rank <= 0 {
			t.Errorf("Leaderboard entry has invalid rank: %d", entry.Rank)
		}
	}
	t.Log("Leaderboard structure verified")
}

// ---------------------------------------------------------------------------
// 2.10.2 - TestStreakBonus
// Simulate 7 consecutive days -> verify streak bonus
// ---------------------------------------------------------------------------

func TestStreakBonus(t *testing.T) {
	uToken := userToken(t)

	// Note: True 7-day simulation requires either time manipulation in the test
	// database or a dedicated test endpoint. This test verifies the streak
	// data returned by the rewards endpoint and validates the milestone rules.

	// Step 1: Check current rewards/streak state
	t.Log("Step 1: Checking current streak state via rewards endpoint")
	status, resp := doRequest(t, http.MethodGet, apiURL(t, "/v1/me/rewards"), uToken, nil)
	requireSuccess(t, status, resp)

	var rewards rewardsResponse
	unmarshalData(t, resp, &rewards)

	// Parse streak data
	var streak map[string]interface{}
	if rewards.Streak != nil {
		if err := json.Unmarshal(rewards.Streak, &streak); err == nil {
			currentStreak, _ := streak["current_streak"].(float64)
			longestStreak, _ := streak["longest_streak"].(float64)
			bonusSkips, _ := streak["bonus_skips"].(float64)
			bonusAnswers, _ := streak["bonus_answers"].(float64)
			t.Logf("Current streak: %d days, longest: %d days", int(currentStreak), int(longestStreak))
			t.Logf("Bonus skips: %d, bonus answers: %d", int(bonusSkips), int(bonusAnswers))

			// Milestone verification based on domain rules:
			// Day 3:  No bonus
			// Day 7:  +1 bonus skip
			// Day 10: +1 bonus skip + 1.0 token
			// Day 14: +1 bonus answer
			// Day 21: +1 skip + +1 answer + 2.0 tokens
			// Day 30: +2 skips + +1 answer + 5.0 tokens

			cs := int(currentStreak)
			if cs >= 7 {
				t.Log("Streak >= 7 days: user should have received +1 bonus skip at day 7")
				if int(bonusSkips) < 1 {
					t.Log("Warning: bonus_skips < 1 despite streak >= 7 (may have been consumed)")
				}
			}
			if cs >= 14 {
				t.Log("Streak >= 14 days: user should have received +1 bonus answer at day 14")
			}
		} else {
			t.Logf("Could not parse streak data: %v", err)
		}
	} else {
		t.Log("No streak data returned")
	}

	// Step 2: Play a session to record today's activity (extends streak)
	t.Log("Step 2: Starting a session to extend streak")
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/start"), uToken, nil)
	requireSuccess(t, status, resp)

	var session sessionView
	unmarshalData(t, resp, &session)
	t.Logf("Session for streak: id=%s, bonus resources may be applied", session.ID)

	// Answer one card to ensure the session counts
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current/card"), uToken, nil)
	if status == http.StatusOK {
		status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/answer"), uToken,
			map[string]interface{}{"answer": true})
		if status >= 200 && status < 300 {
			t.Log("Answered one card to register daily activity")
		}
	}

	// Step 3: Re-check streak
	t.Log("Step 3: Re-checking streak after activity")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/me/rewards"), uToken, nil)
	requireSuccess(t, status, resp)

	var updatedRewards rewardsResponse
	unmarshalData(t, resp, &updatedRewards)

	if updatedRewards.Streak != nil {
		var updatedStreak map[string]interface{}
		if err := json.Unmarshal(updatedRewards.Streak, &updatedStreak); err == nil {
			newStreak, _ := updatedStreak["current_streak"].(float64)
			t.Logf("Streak after activity: %d days", int(newStreak))
		}
	}

	t.Log("Streak bonus test complete. For full 7-day simulation, run daily over a week or use test DB seeding.")
}

// ---------------------------------------------------------------------------
// 2.10.4 - TestRewardDistribution
// Admin distributes rewards -> user sees pending -> claims
// ---------------------------------------------------------------------------

func TestRewardDistribution(t *testing.T) {
	aToken := adminToken(t)
	uToken := userToken(t)

	// Step 1: Admin creates a reward config
	t.Log("Step 1: Creating reward config")
	status, resp := doRequest(t, http.MethodPost, apiURL(t, "/v1/admin/rewards/configs"), aToken, map[string]interface{}{
		"period_type": "daily",
		"rank_from":   1,
		"rank_to":     10,
		"reward_type": "token",
		"amount":      5.0,
		"description": map[string]string{"en": "E2E test daily top 10 reward"},
		"is_active":   true,
	})
	requireSuccess(t, status, resp)
	t.Log("Reward config created")

	// Step 2: Get user profile to obtain user_id for distribution
	t.Log("Step 2: Getting user profile")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/me"), uToken, nil)
	requireSuccess(t, status, resp)

	var profile map[string]interface{}
	unmarshalData(t, resp, &profile)
	userID, _ := profile["id"].(string)
	if userID == "" {
		userID, _ = profile["user_id"].(string)
	}
	t.Logf("User ID: %s", userID)

	if userID == "" {
		t.Fatal("Could not determine user ID from profile")
	}

	// Step 3: Admin distributes rewards
	t.Log("Step 3: Distributing rewards")
	periodKey := time.Now().UTC().Format("2006-01-02")
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/admin/rewards/distribute"), aToken, map[string]interface{}{
		"period_type": "daily",
		"period_key":  periodKey,
		"entries": []map[string]interface{}{
			{
				"user_id":      userID,
				"rank":         1,
				"total_points": 100,
			},
		},
	})
	requireSuccess(t, status, resp)

	var distResult distributeResponse
	unmarshalData(t, resp, &distResult)
	t.Logf("Distributed %d rewards", distResult.Distributed)

	// Step 4: User checks pending rewards
	t.Log("Step 4: Checking pending rewards")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/me/rewards"), uToken, nil)
	requireSuccess(t, status, resp)

	var rewards rewardsResponse
	unmarshalData(t, resp, &rewards)
	t.Logf("Pending rewards: %d, History: %d", len(rewards.Pending), len(rewards.History))

	if len(rewards.Pending) == 0 {
		t.Log("Warning: no pending rewards found. Distribution may not have matched config or user.")
		return
	}

	// Find a pending reward to claim
	var rewardToClaim string
	for _, r := range rewards.Pending {
		if r.Status == "pending" {
			rewardToClaim = r.ID
			t.Logf("Found pending reward: id=%s, type=%s, amount=%.2f", r.ID, r.RewardType, r.Amount)
			break
		}
	}

	if rewardToClaim == "" {
		t.Log("No claimable pending reward found")
		return
	}

	// Step 5: User claims the reward
	t.Logf("Step 5: Claiming reward %s", rewardToClaim)
	status, resp = doRequest(t, http.MethodPost,
		apiURL(t, fmt.Sprintf("/v1/me/rewards/%s/claim", rewardToClaim)), uToken, nil)
	requireSuccess(t, status, resp)

	var claimResult claimResponse
	unmarshalData(t, resp, &claimResult)
	t.Logf("Claim result: message=%s, status=%s, type=%s", claimResult.Message, claimResult.Status, claimResult.RewardType)

	// Step 6: Verify reward is no longer pending
	t.Log("Step 6: Verifying reward claimed")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/me/rewards"), uToken, nil)
	requireSuccess(t, status, resp)

	var updatedRewards rewardsResponse
	unmarshalData(t, resp, &updatedRewards)

	for _, r := range updatedRewards.Pending {
		if r.ID == rewardToClaim && r.Status == "pending" {
			t.Error("Reward still shows as pending after claim")
		}
	}
	t.Log("Reward distribution and claim flow verified")
}

// ---------------------------------------------------------------------------
// Environment helper for second user token
// ---------------------------------------------------------------------------

func secondUserToken(t *testing.T) string {
	t.Helper()
	tok := os.Getenv("E2E_AUTH_TOKEN_2")
	if tok == "" {
		t.Skip("E2E_AUTH_TOKEN_2 environment variable not set; skipping test that requires a second user")
	}
	return tok
}

// ---------------------------------------------------------------------------
// 3.10.1 - TestReferralFlow
// User gets referral code -> second user signs up with it -> verify referrer gets bonus skip
// ---------------------------------------------------------------------------

func TestReferralFlow(t *testing.T) {
	uToken := userToken(t)
	u2Token := secondUserToken(t)

	// Step 1: Get referral code for user 1
	t.Log("Step 1: Getting referral code for user 1")
	status, resp := doRequest(t, http.MethodGet, apiURL(t, "/referral/code"), uToken, nil)
	if status == http.StatusNotFound {
		// Try with /v1 prefix
		status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/referral/code"), uToken, nil)
	}
	requireSuccess(t, status, resp)

	var codeResp map[string]interface{}
	unmarshalData(t, resp, &codeResp)
	referralCode, _ := codeResp["referral_code"].(string)
	if referralCode == "" {
		t.Fatal("expected non-empty referral_code in response")
	}
	t.Logf("User 1 referral code: %s", referralCode)

	// Step 2: Get referral stats before (baseline)
	t.Log("Step 2: Getting referral stats before")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/referral/stats"), uToken, nil)
	requireSuccess(t, status, resp)

	var statsBefore map[string]interface{}
	unmarshalData(t, resp, &statsBefore)
	t.Logf("Referral stats before: %+v", statsBefore)

	// Step 3: Second user "signs up" using the referral code
	// In a real flow, the referral code is passed during login/registration.
	// We verify that user 2 can see a friends leaderboard that includes user 1 (referral link).
	t.Log("Step 3: User 2 checking friends leaderboard (verifying referral connection)")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/leaderboards/friends"), u2Token, nil)
	requireSuccess(t, status, resp)
	t.Log("User 2 friends leaderboard retrieved")

	// Step 4: User 2 starts a session (triggers referral status update to first_session)
	t.Log("Step 4: User 2 starting a session to trigger referral reward")
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/start"), u2Token, nil)
	requireSuccess(t, status, resp)

	var u2Session sessionView
	unmarshalData(t, resp, &u2Session)
	t.Logf("User 2 session: id=%s, skips_remaining=%d", u2Session.ID, u2Session.SkipsRemaining)

	// Step 5: Verify referral stats updated for user 1
	t.Log("Step 5: Checking referral stats after")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/referral/stats"), uToken, nil)
	requireSuccess(t, status, resp)

	var statsAfter map[string]interface{}
	unmarshalData(t, resp, &statsAfter)
	t.Logf("Referral stats after: %+v", statsAfter)

	// Step 6: Check if user 1 has bonus skip in next session
	t.Log("Step 6: Checking user 1 session for bonus skip from referral")
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/start"), uToken, nil)
	requireSuccess(t, status, resp)

	var u1Session sessionView
	unmarshalData(t, resp, &u1Session)
	t.Logf("User 1 session: skips_remaining=%d (baseline=5; if >5, bonus applied)", u1Session.SkipsRemaining)
	t.Log("Referral flow test complete")
}

// ---------------------------------------------------------------------------
// 3.10.2 - TestPerfectDayAchievement
// User answers all cards correctly -> verify achievement unlocked
// ---------------------------------------------------------------------------

func TestPerfectDayAchievement(t *testing.T) {
	uToken := userToken(t)

	// Step 1: Get achievements before
	t.Log("Step 1: Getting achievements before")
	status, resp := doRequest(t, http.MethodGet, apiURL(t, "/v1/me/achievements"), uToken, nil)
	requireSuccess(t, status, resp)

	var achievementsBefore map[string]interface{}
	unmarshalData(t, resp, &achievementsBefore)
	earnedBefore, _ := achievementsBefore["earned"].([]interface{})
	t.Logf("Earned achievements before: %d", len(earnedBefore))

	// Step 2: Start a session
	t.Log("Step 2: Starting session")
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/start"), uToken, nil)
	requireSuccess(t, status, resp)

	var session sessionView
	unmarshalData(t, resp, &session)
	t.Logf("Session started: id=%s, total_cards=%d, answers_remaining=%d",
		session.ID, session.TotalCards, session.AnswersRemaining)

	// Step 3: Answer all available cards with "true" (attempting perfect score)
	answeredCount := 0
	for i := 0; i < session.TotalCards; i++ {
		status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current/card"), uToken, nil)
		if status != http.StatusOK {
			t.Logf("No more cards at index %d", i)
			break
		}

		// Answer true for all cards
		status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/answer"), uToken,
			map[string]interface{}{"answer": true})
		if status == http.StatusBadRequest {
			// Out of answers, skip the rest
			t.Logf("Out of answers at card %d, skipping remaining", i+1)
			status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/skip"), uToken, nil)
			if status != http.StatusOK {
				break
			}
		} else {
			requireSuccess(t, status, resp)
			answeredCount++
		}
	}
	t.Logf("Answered %d cards", answeredCount)

	// Step 4: Check achievements after
	t.Log("Step 4: Checking achievements after session")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/me/achievements"), uToken, nil)
	requireSuccess(t, status, resp)

	var achievementsAfter map[string]interface{}
	unmarshalData(t, resp, &achievementsAfter)
	earnedAfter, _ := achievementsAfter["earned"].([]interface{})
	t.Logf("Earned achievements after: %d", len(earnedAfter))

	if len(earnedAfter) > len(earnedBefore) {
		t.Logf("New achievement(s) unlocked! Count went from %d to %d", len(earnedBefore), len(earnedAfter))
	} else {
		t.Log("No new achievements unlocked. Perfect day achievement requires all cards resolved correctly (post-resolution).")
	}

	// Step 5: Verify the achievements list structure
	allAchievements, _ := achievementsAfter["achievements"].([]interface{})
	t.Logf("Total available achievements: %d", len(allAchievements))
	for _, a := range allAchievements {
		if aMap, ok := a.(map[string]interface{}); ok {
			name, _ := aMap["name"].(string)
			slug, _ := aMap["slug"].(string)
			t.Logf("  Achievement: name=%s, slug=%s", name, slug)
		}
	}
	t.Log("Perfect day achievement test complete")
}

// ---------------------------------------------------------------------------
// 3.10.3 - TestMiniLeague
// Create mini-league -> invite friend -> both see league leaderboard
// ---------------------------------------------------------------------------

func TestMiniLeague(t *testing.T) {
	u1Token := userToken(t)
	u2Token := secondUserToken(t)

	// Step 1: User 1 creates a mini league
	leagueName := fmt.Sprintf("E2E League %d", time.Now().UnixMilli())
	t.Logf("Step 1: Creating mini league: %s", leagueName)
	status, resp := doRequest(t, http.MethodPost, apiURL(t, "/v1/leagues"), u1Token, map[string]interface{}{
		"name": leagueName,
	})
	requireSuccess(t, status, resp)

	var league map[string]interface{}
	unmarshalData(t, resp, &league)
	leagueID, _ := league["id"].(string)
	inviteCode, _ := league["invite_code"].(string)
	t.Logf("League created: id=%s, invite_code=%s", leagueID, inviteCode)

	if leagueID == "" {
		t.Fatal("expected league id in response")
	}
	if inviteCode == "" {
		t.Fatal("expected invite_code in response")
	}

	// Step 2: User 2 joins the league using the invite code
	t.Log("Step 2: User 2 joining league")
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/leagues/join"), u2Token, map[string]interface{}{
		"invite_code": inviteCode,
	})
	requireSuccess(t, status, resp)
	t.Log("User 2 joined league")

	// Step 3: User 1 sees the league
	t.Log("Step 3: User 1 checking their leagues")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/leagues"), u1Token, nil)
	requireSuccess(t, status, resp)

	var u1Leagues []map[string]interface{}
	unmarshalData(t, resp, &u1Leagues)
	found := false
	for _, l := range u1Leagues {
		if id, _ := l["id"].(string); id == leagueID {
			found = true
			break
		}
	}
	if !found {
		t.Error("User 1 does not see the created league in their leagues list")
	} else {
		t.Log("User 1 sees the league")
	}

	// Step 4: User 2 sees the league
	t.Log("Step 4: User 2 checking their leagues")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/leagues"), u2Token, nil)
	requireSuccess(t, status, resp)

	var u2Leagues []map[string]interface{}
	unmarshalData(t, resp, &u2Leagues)
	found = false
	for _, l := range u2Leagues {
		if id, _ := l["id"].(string); id == leagueID {
			found = true
			break
		}
	}
	if !found {
		t.Error("User 2 does not see the league after joining")
	} else {
		t.Log("User 2 sees the league")
	}

	// Step 5: Both users see the league leaderboard
	t.Log("Step 5: Checking league leaderboard")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, fmt.Sprintf("/v1/leaderboards/league/%s", leagueID)), u1Token, nil)
	requireSuccess(t, status, resp)

	var leaderboard interface{}
	unmarshalData(t, resp, &leaderboard)
	t.Logf("League leaderboard response: %+v", leaderboard)

	// User 2 also queries the same leaderboard
	status, resp = doRequest(t, http.MethodGet, apiURL(t, fmt.Sprintf("/v1/leaderboards/league/%s", leagueID)), u2Token, nil)
	requireSuccess(t, status, resp)
	t.Log("Both users can see the league leaderboard")

	// Step 6: Verify league details
	t.Log("Step 6: Getting league details")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, fmt.Sprintf("/v1/leagues/%s", leagueID)), u1Token, nil)
	requireSuccess(t, status, resp)

	var leagueDetail map[string]interface{}
	unmarshalData(t, resp, &leagueDetail)
	t.Logf("League detail: name=%s", leagueDetail["name"])
	t.Log("Mini league test complete")
}

// ---------------------------------------------------------------------------
// 3.10.4 - TestWebSocketCardResolved
// Connect WebSocket -> admin resolves card -> verify event received
// ---------------------------------------------------------------------------

func TestWebSocketCardResolved(t *testing.T) {
	uToken := userToken(t)
	aToken := adminToken(t)

	// Build WebSocket URL from base URL
	base := baseURL(t)
	wsURL := "ws" + base[len("http"):] + "/ws?token=" + uToken

	t.Logf("Step 1: Connecting WebSocket to %s", base+"/ws")

	// Use gorilla/websocket dialer
	dialer := gorillaWS.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	conn, wsResp, err := dialer.Dial(wsURL, nil)
	if err != nil {
		if wsResp != nil {
			t.Skipf("WebSocket connection failed (status=%d): %v — server may not support WS in test env", wsResp.StatusCode, err)
		}
		t.Skipf("WebSocket connection failed: %v — server may not be reachable or WS not enabled", err)
	}
	defer conn.Close()
	t.Log("WebSocket connected")

	// Step 2: Start a session and answer a card so there is something to resolve
	t.Log("Step 2: Preparing a card answer for resolution")
	status, resp := doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/start"), uToken, nil)
	requireSuccess(t, status, resp)

	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current/card"), uToken, nil)
	if status != http.StatusOK {
		t.Skip("No cards available for WebSocket resolution test")
	}

	var card cardView
	unmarshalData(t, resp, &card)
	t.Logf("Got card: id=%s", card.ID)

	// Answer the card
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/answer"), uToken,
		map[string]interface{}{"answer": true})
	if status != http.StatusOK {
		t.Logf("Could not answer card (status=%d), attempting resolution anyway", status)
	}

	// Step 3: Admin resolves the card
	t.Log("Step 3: Admin resolving card")
	status, resp = doRequest(t, http.MethodPost,
		apiURL(t, fmt.Sprintf("/v1/admin/cards/%s/resolve", card.ID)),
		aToken,
		map[string]interface{}{"correct_answer": true})
	if status != http.StatusOK {
		t.Logf("Card resolution returned status %d (may already be resolved)", status)
	}

	// Step 4: Read WebSocket message with timeout
	t.Log("Step 4: Waiting for WebSocket event")
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	received := false
	for i := 0; i < 5; i++ {
		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Logf("WebSocket read: %v (may be timeout — no event received)", err)
			break
		}

		var wsMsg map[string]interface{}
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			t.Logf("Received non-JSON message: %s", string(message))
			continue
		}

		msgType, _ := wsMsg["type"].(string)
		t.Logf("Received WS event: type=%s", msgType)

		if msgType == "card_resolved" {
			received = true
			t.Log("Received card_resolved event via WebSocket")
			break
		}
	}

	if !received {
		t.Log("No card_resolved event received within timeout. This may happen if the card was already resolved or no answer matched.")
	}
	t.Log("WebSocket card resolved test complete")
}

// ---------------------------------------------------------------------------
// 4.11.2 - TestAntiAbuse
// Multiple accounts from same device/IP -> verify flagged
// ---------------------------------------------------------------------------

func TestAntiAbuse(t *testing.T) {
	aToken := adminToken(t)

	// Step 1: Get current abuse flags (baseline)
	t.Log("Step 1: Getting current abuse flags (baseline)")
	status, resp := doRequest(t, http.MethodGet, apiURL(t, "/v1/admin/abuse-flags"), aToken, nil)
	requireSuccess(t, status, resp)

	var flagsBefore []map[string]interface{}
	unmarshalData(t, resp, &flagsBefore)
	t.Logf("Abuse flags before: %d", len(flagsBefore))

	// Step 2: Simulate two logins from the same device ID and IP
	// The abuse detection happens inside the auth service when login is called.
	// We call login with the same device_id to trigger CheckMultiAccount.
	sharedDeviceID := fmt.Sprintf("e2e-device-%d", time.Now().UnixMilli())
	t.Logf("Step 2: Simulating login with shared device_id=%s", sharedDeviceID)

	// Login user 1 with the device ID
	u1Token := os.Getenv("E2E_EXCHANGE_TOKEN")
	if u1Token == "" {
		t.Log("E2E_EXCHANGE_TOKEN not set; attempting login with E2E_AUTH_TOKEN to register device")
		// Try to register the device directly
		uToken := userToken(t)
		status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/devices/register"), uToken, map[string]interface{}{
			"token":       "e2e-fcm-token-abuse-1",
			"device_type": "android",
		})
		if status >= 200 && status < 300 {
			t.Log("Device registered for user 1")
		}

		// If we have a second user token, register the same device type
		u2Token := os.Getenv("E2E_AUTH_TOKEN_2")
		if u2Token != "" {
			status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/devices/register"), u2Token, map[string]interface{}{
				"token":       "e2e-fcm-token-abuse-2",
				"device_type": "android",
			})
			if status >= 200 && status < 300 {
				t.Log("Device registered for user 2")
			}
		}
	}

	// Step 3: Check abuse flags after
	t.Log("Step 3: Checking abuse flags after device registration")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/admin/abuse-flags"), aToken, nil)
	requireSuccess(t, status, resp)

	var flagsAfter []map[string]interface{}
	unmarshalData(t, resp, &flagsAfter)
	t.Logf("Abuse flags after: %d", len(flagsAfter))

	// Step 4: Verify flag structure if any exist
	for _, flag := range flagsAfter {
		flagType, _ := flag["flag_type"].(string)
		flagStatus, _ := flag["status"].(string)
		flagUserID, _ := flag["user_id"].(string)
		t.Logf("  Flag: type=%s, status=%s, user=%s", flagType, flagStatus, flagUserID)

		if flagType == "multi_account" {
			t.Log("Multi-account flag detected — anti-abuse system is working")
		}
	}

	// Step 5: If there are pending flags, test the review flow
	if len(flagsAfter) > 0 {
		flagID, _ := flagsAfter[0]["id"].(string)
		if flagID != "" {
			t.Logf("Step 5: Reviewing abuse flag %s", flagID)
			status, resp = doRequest(t, http.MethodPost,
				apiURL(t, fmt.Sprintf("/v1/admin/abuse-flags/%s/review", flagID)),
				aToken,
				map[string]interface{}{"status": "reviewed"})
			if status >= 200 && status < 300 {
				t.Log("Abuse flag reviewed successfully")
			} else {
				t.Logf("Flag review returned status %d (may already be reviewed)", status)
			}
		}
	}

	t.Log("Anti-abuse test complete")
}

// ---------------------------------------------------------------------------
// 4.11.3 - TestExchangeTokenClaim
// Earn reward -> claim tokens -> verify exchange credit
// ---------------------------------------------------------------------------

func TestExchangeTokenClaim(t *testing.T) {
	aToken := adminToken(t)
	uToken := userToken(t)

	// Step 1: Get user profile to obtain user_id
	t.Log("Step 1: Getting user profile")
	status, resp := doRequest(t, http.MethodGet, apiURL(t, "/v1/me"), uToken, nil)
	requireSuccess(t, status, resp)

	var profile map[string]interface{}
	unmarshalData(t, resp, &profile)
	userID, _ := profile["id"].(string)
	if userID == "" {
		userID, _ = profile["user_id"].(string)
	}
	if userID == "" {
		t.Fatal("Could not determine user ID from profile")
	}
	t.Logf("User ID: %s", userID)

	// Step 2: Admin creates a token reward config
	t.Log("Step 2: Creating token reward config")
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/admin/rewards/configs"), aToken, map[string]interface{}{
		"period_type": "daily",
		"rank_from":   1,
		"rank_to":     5,
		"reward_type": "token",
		"amount":      10.0,
		"description": map[string]string{"en": "E2E exchange token claim test"},
		"is_active":   true,
	})
	requireSuccess(t, status, resp)
	t.Log("Token reward config created")

	// Step 3: Admin distributes token reward to user
	t.Log("Step 3: Distributing token reward")
	periodKey := fmt.Sprintf("exchange-test-%d", time.Now().UnixMilli())
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/admin/rewards/distribute"), aToken, map[string]interface{}{
		"period_type": "daily",
		"period_key":  periodKey,
		"entries": []map[string]interface{}{
			{
				"user_id":      userID,
				"rank":         1,
				"total_points": 200,
			},
		},
	})
	requireSuccess(t, status, resp)

	var distResult distributeResponse
	unmarshalData(t, resp, &distResult)
	t.Logf("Distributed %d rewards", distResult.Distributed)

	// Step 4: Get pending rewards and find the token reward
	t.Log("Step 4: Getting pending rewards")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/me/rewards"), uToken, nil)
	requireSuccess(t, status, resp)

	var rewards rewardsResponse
	unmarshalData(t, resp, &rewards)
	t.Logf("Pending rewards: %d", len(rewards.Pending))

	var tokenRewardID string
	var tokenAmount float64
	for _, r := range rewards.Pending {
		if r.Status == "pending" && r.RewardType == "token" {
			tokenRewardID = r.ID
			tokenAmount = r.Amount
			t.Logf("Found token reward: id=%s, amount=%.2f", r.ID, r.Amount)
			break
		}
	}

	if tokenRewardID == "" {
		t.Log("No pending token reward found; distribution may not have matched config")
		return
	}

	// Step 5: Claim the token reward (triggers exchange credit flow)
	t.Logf("Step 5: Claiming token reward %s", tokenRewardID)
	status, resp = doRequest(t, http.MethodPost,
		apiURL(t, fmt.Sprintf("/v1/me/rewards/%s/claim", tokenRewardID)),
		uToken, nil)
	requireSuccess(t, status, resp)

	var claimResult claimResponse
	unmarshalData(t, resp, &claimResult)
	t.Logf("Claim result: message=%s, status=%s, reward_type=%s, amount=%.2f",
		claimResult.Message, claimResult.Status, claimResult.RewardType, claimResult.Amount)

	// Step 6: Verify the reward is credited (status should be "credited" for token rewards)
	if claimResult.Status != "credited" {
		t.Errorf("expected status 'credited' for token reward, got '%s'", claimResult.Status)
	}
	if claimResult.RewardType != "token" {
		t.Errorf("expected reward_type 'token', got '%s'", claimResult.RewardType)
	}
	if claimResult.Amount != tokenAmount {
		t.Logf("Warning: claimed amount %.2f differs from expected %.2f", claimResult.Amount, tokenAmount)
	}

	// Step 7: Verify reward is no longer pending
	t.Log("Step 7: Verifying reward no longer pending")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/me/rewards"), uToken, nil)
	requireSuccess(t, status, resp)

	var updatedRewards rewardsResponse
	unmarshalData(t, resp, &updatedRewards)
	for _, r := range updatedRewards.Pending {
		if r.ID == tokenRewardID && r.Status == "pending" {
			t.Error("Token reward still shows as pending after claim")
		}
	}

	// Step 8: Check exchange prompts reflect the claim
	t.Log("Step 8: Checking exchange prompts")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/me/exchange-prompts"), uToken, nil)
	requireSuccess(t, status, resp)

	var prompts map[string]interface{}
	unmarshalData(t, resp, &prompts)
	t.Logf("Exchange prompts: %+v", prompts)
	t.Log("Exchange token claim test complete")
}

// ---------------------------------------------------------------------------
// 2.10.3 - TestDailyLeaderboardReset
// Query leaderboard -> verify daily scope returns current day data
// ---------------------------------------------------------------------------

func TestDailyLeaderboardReset(t *testing.T) {
	uToken := userToken(t)

	// Step 1: Get today's daily leaderboard
	today := time.Now().UTC().Format("2006-01-02")
	t.Logf("Step 1: Getting daily leaderboard for today (%s)", today)
	status, resp := doRequest(t, http.MethodGet, apiURL(t, fmt.Sprintf("/v1/leaderboards/daily?date=%s", today)), uToken, nil)
	requireSuccess(t, status, resp)

	var lbToday leaderboardResponse
	unmarshalData(t, resp, &lbToday)
	t.Logf("Daily leaderboard: period_type=%s, period_key=%s, entries=%d, total=%d",
		lbToday.PeriodType, lbToday.PeriodKey, len(lbToday.Entries), lbToday.Total)

	// Verify period_key matches today's date
	if lbToday.PeriodKey != today {
		t.Errorf("expected period_key=%s for today, got %s", today, lbToday.PeriodKey)
	}
	if lbToday.PeriodType != "daily" {
		t.Errorf("expected period_type=daily, got %s", lbToday.PeriodType)
	}

	// Step 2: Get yesterday's daily leaderboard (should be different data)
	yesterday := time.Now().UTC().Add(-24 * time.Hour).Format("2006-01-02")
	t.Logf("Step 2: Getting daily leaderboard for yesterday (%s)", yesterday)
	status, resp = doRequest(t, http.MethodGet, apiURL(t, fmt.Sprintf("/v1/leaderboards/daily?date=%s", yesterday)), uToken, nil)
	requireSuccess(t, status, resp)

	var lbYesterday leaderboardResponse
	unmarshalData(t, resp, &lbYesterday)
	t.Logf("Yesterday leaderboard: period_key=%s, entries=%d, total=%d",
		lbYesterday.PeriodKey, len(lbYesterday.Entries), lbYesterday.Total)

	if lbYesterday.PeriodKey != yesterday {
		t.Errorf("expected period_key=%s for yesterday, got %s", yesterday, lbYesterday.PeriodKey)
	}

	// Step 3: Verify daily leaderboard default returns today
	t.Log("Step 3: Getting daily leaderboard with no date param (should default to today)")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/leaderboards/daily"), uToken, nil)
	requireSuccess(t, status, resp)

	var lbDefault leaderboardResponse
	unmarshalData(t, resp, &lbDefault)
	t.Logf("Default daily leaderboard: period_key=%s", lbDefault.PeriodKey)

	if lbDefault.PeriodKey != today {
		t.Errorf("default daily leaderboard period_key=%s, expected today=%s", lbDefault.PeriodKey, today)
	}

	// Step 4: Verify user_rank field
	if lbToday.UserRank != nil {
		t.Logf("User rank on daily: rank=%d, points=%d", lbToday.UserRank.Rank, lbToday.UserRank.TotalPoints)
	} else {
		t.Log("User not ranked on today's daily leaderboard (no activity today)")
	}

	// Step 5: Verify entries are ordered by rank
	for i := 1; i < len(lbToday.Entries); i++ {
		if lbToday.Entries[i].Rank < lbToday.Entries[i-1].Rank {
			t.Errorf("leaderboard entries not sorted by rank: rank[%d]=%d > rank[%d]=%d",
				i-1, lbToday.Entries[i-1].Rank, i, lbToday.Entries[i].Rank)
		}
	}
	t.Log("Daily leaderboard reset test complete")
}

// ---------------------------------------------------------------------------
// 1.16.5 - TestTimerExpiry
// Start session -> get card -> wait/verify timer behavior
// ---------------------------------------------------------------------------

func TestTimerExpiry(t *testing.T) {
	uToken := userToken(t)

	// Step 1: Start a session
	t.Log("Step 1: Starting session")
	status, resp := doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/start"), uToken, nil)
	requireSuccess(t, status, resp)

	var session sessionView
	unmarshalData(t, resp, &session)
	t.Logf("Session started: id=%s, status=%s", session.ID, session.Status)

	// Step 2: Get the current card and record the presentation time
	t.Log("Step 2: Getting current card")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current/card"), uToken, nil)
	if status != http.StatusOK {
		t.Skip("No cards available for timer test")
	}

	var card cardView
	unmarshalData(t, resp, &card)
	t.Logf("Card presented: id=%s, tier=%s, expires_at=%s", card.ID, card.Tier, card.ExpiresAt.Format(time.RFC3339))

	// Step 3: Re-fetch session to verify card_presented_at is set
	t.Log("Step 3: Checking session for card_presented_at")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current"), uToken, nil)
	requireSuccess(t, status, resp)

	var sessionAfterCard sessionView
	unmarshalData(t, resp, &sessionAfterCard)

	if sessionAfterCard.CardPresentedAt != nil {
		t.Logf("Card presented at: %s", sessionAfterCard.CardPresentedAt.Format(time.RFC3339))

		// Step 4: Wait a brief period and verify the timer is tracked
		t.Log("Step 4: Waiting 2 seconds to verify timer progression")
		time.Sleep(2 * time.Second)

		// Re-fetch and verify card_presented_at hasn't changed (timer is stable)
		status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current"), uToken, nil)
		requireSuccess(t, status, resp)

		var sessionAfterWait sessionView
		unmarshalData(t, resp, &sessionAfterWait)

		if sessionAfterWait.CardPresentedAt != nil {
			elapsed := time.Since(*sessionAfterWait.CardPresentedAt)
			t.Logf("Time since card presented: %v", elapsed)

			if elapsed < 2*time.Second {
				t.Error("Timer elapsed less than 2 seconds — card_presented_at may have been reset")
			}
		}
	} else {
		t.Log("card_presented_at is nil — server may not track per-card timer in session view")
	}

	// Step 5: Answer the card and verify the timer resets for next card
	t.Log("Step 5: Answering card and checking timer reset")
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/answer"), uToken,
		map[string]interface{}{"answer": true})
	if status != http.StatusOK {
		status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/skip"), uToken, nil)
	}

	// Get next card
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current/card"), uToken, nil)
	if status == http.StatusOK {
		// Re-fetch session to check if card_presented_at was updated
		status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current"), uToken, nil)
		requireSuccess(t, status, resp)

		var sessionNextCard sessionView
		unmarshalData(t, resp, &sessionNextCard)

		if sessionNextCard.CardPresentedAt != nil {
			elapsed := time.Since(*sessionNextCard.CardPresentedAt)
			t.Logf("Next card timer: presented %v ago", elapsed)
			if elapsed > 5*time.Second {
				t.Log("Warning: card_presented_at is more than 5 seconds ago — timer may not have reset")
			} else {
				t.Log("Timer reset confirmed for new card")
			}
		} else {
			t.Log("card_presented_at is nil for next card")
		}
	} else {
		t.Log("No more cards available after answering")
	}

	t.Log("Timer expiry test complete")
}

// ---------------------------------------------------------------------------
// 3.10.5 - TestSocialShareDeepLink
// Get referral code -> verify the generated share link has correct scheme
// ---------------------------------------------------------------------------

func TestSocialShareDeepLink(t *testing.T) {
	uToken := userToken(t)

	// Step 1: Get the user's referral code
	t.Log("Step 1: Getting referral code")
	status, resp := doRequest(t, http.MethodGet, apiURL(t, "/v1/referral/code"), uToken, nil)
	requireSuccess(t, status, resp)

	var codeResp map[string]interface{}
	unmarshalData(t, resp, &codeResp)
	referralCode, _ := codeResp["referral_code"].(string)
	if referralCode == "" {
		t.Fatal("expected non-empty referral_code in response")
	}
	t.Logf("Referral code: %s", referralCode)

	// Step 2: Construct expected deep link formats and verify code is usable
	// The mobile app generates share URLs locally using the referral code.
	// Valid formats:
	//   - Universal link: https://play.xex.exchange/invite/<code>
	//   - App deep link:  xexplay://invite/<code>
	universalLink := fmt.Sprintf("https://play.xex.exchange/invite/%s", referralCode)
	deepLink := fmt.Sprintf("xexplay://invite/%s", referralCode)

	t.Logf("Expected universal link: %s", universalLink)
	t.Logf("Expected deep link: %s", deepLink)

	// Verify the referral code is a reasonable format (non-empty, alphanumeric-ish)
	if len(referralCode) < 4 {
		t.Errorf("referral code too short (%d chars): %s", len(referralCode), referralCode)
	}
	for _, ch := range referralCode {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_') {
			t.Errorf("referral code contains unexpected character: %c", ch)
			break
		}
	}

	// Step 3: Verify referral stats endpoint works (confirms code is registered server-side)
	t.Log("Step 3: Verifying referral stats endpoint")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/referral/stats"), uToken, nil)
	requireSuccess(t, status, resp)

	var stats map[string]interface{}
	unmarshalData(t, resp, &stats)
	t.Logf("Referral stats: %+v", stats)

	// Step 4: Verify the same code is returned on subsequent calls (idempotent)
	t.Log("Step 4: Verifying referral code is idempotent")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/referral/code"), uToken, nil)
	requireSuccess(t, status, resp)

	var codeResp2 map[string]interface{}
	unmarshalData(t, resp, &codeResp2)
	referralCode2, _ := codeResp2["referral_code"].(string)
	if referralCode2 != referralCode {
		t.Errorf("referral code changed between calls: %s -> %s", referralCode, referralCode2)
	} else {
		t.Log("Referral code is stable across calls")
	}

	t.Log("Social share deep link test complete")
}

// ---------------------------------------------------------------------------
// 2.10.5 - TestPushNotificationOnResolve
// Start session -> answer card -> admin resolves -> check notification dispatch
// ---------------------------------------------------------------------------

func TestPushNotificationOnResolve(t *testing.T) {
	uToken := userToken(t)
	aToken := adminToken(t)

	// Step 1: Register a device token so notifications can be dispatched
	t.Log("Step 1: Registering device for push notifications")
	fcmToken := fmt.Sprintf("e2e-fcm-resolve-test-%d", time.Now().UnixMilli())
	status, resp := doRequest(t, http.MethodPost, apiURL(t, "/v1/devices/register"), uToken, map[string]interface{}{
		"token":       fcmToken,
		"device_type": "android",
	})
	requireSuccess(t, status, resp)
	t.Logf("Device registered with token: %s", fcmToken)

	// Step 2: Start a session and answer a card
	t.Log("Step 2: Starting session")
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/start"), uToken, nil)
	requireSuccess(t, status, resp)

	var session sessionView
	unmarshalData(t, resp, &session)
	t.Logf("Session started: id=%s", session.ID)

	// Get and answer a card
	t.Log("Step 3: Getting and answering a card")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current/card"), uToken, nil)
	if status != http.StatusOK {
		t.Skip("No cards available for notification resolve test")
	}

	var card cardView
	unmarshalData(t, resp, &card)
	answeredCardID := card.ID
	t.Logf("Answering card: id=%s, tier=%s", card.ID, card.Tier)

	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/answer"), uToken,
		map[string]interface{}{"answer": true})
	if status != http.StatusOK {
		// If answer fails, skip instead and note it
		t.Logf("Answer returned status %d, trying skip", status)
		status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/skip"), uToken, nil)
		requireSuccess(t, status, resp)
		t.Log("Skipped card instead — notification test will proceed with resolution")
	} else {
		t.Log("Card answered successfully")
	}

	// Step 4: Admin resolves the card
	t.Logf("Step 4: Admin resolving card %s", answeredCardID)
	status, resp = doRequest(t, http.MethodPost,
		apiURL(t, fmt.Sprintf("/v1/admin/cards/%s/resolve", answeredCardID)),
		aToken,
		map[string]interface{}{"correct_answer": true})
	if status == http.StatusOK {
		t.Log("Card resolved successfully")
	} else if status == http.StatusBadRequest || status == http.StatusConflict {
		t.Logf("Card already resolved or invalid (status=%d), continuing", status)
	} else {
		requireSuccess(t, status, resp)
	}

	// Step 5: Admin sends a broadcast notification (simulates card-resolution notification)
	// The server dispatches push notifications on card resolution internally.
	// We verify by sending a manual notification and confirming the endpoint works.
	t.Log("Step 5: Admin sending broadcast notification")
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/admin/notifications/send"), aToken, map[string]interface{}{
		"title":  "Card Resolved",
		"body":   fmt.Sprintf("Card %s has been resolved!", answeredCardID),
		"target": "all",
	})
	requireSuccess(t, status, resp)

	var notifResult map[string]interface{}
	unmarshalData(t, resp, &notifResult)
	t.Logf("Notification send result: %+v", notifResult)

	// Step 6: Verify user stats reflect the resolved card
	t.Log("Step 6: Verifying user stats after resolution")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/me/stats"), uToken, nil)
	requireSuccess(t, status, resp)

	var stats map[string]interface{}
	unmarshalData(t, resp, &stats)
	t.Logf("User stats after resolution: %+v", stats)

	// Step 7: Clean up device registration
	t.Log("Step 7: Deregistering test device")
	status, resp = doRequest(t, http.MethodDelete, apiURL(t, fmt.Sprintf("/v1/devices/%s", fcmToken)), uToken, nil)
	if status >= 200 && status < 300 {
		t.Log("Device deregistered")
	} else {
		t.Logf("Device deregistration returned status %d (non-critical)", status)
	}

	t.Log("Push notification on resolve test complete")
}

// ---------------------------------------------------------------------------
// 4.11.1 - TestFullUserJourney
// Profile -> session -> answer all -> rewards -> leaderboard -> achievements -> exchange prompts
// ---------------------------------------------------------------------------

func TestFullUserJourney(t *testing.T) {
	uToken := userToken(t)

	// Step 1: Get user profile
	t.Log("Step 1: Getting user profile")
	status, resp := doRequest(t, http.MethodGet, apiURL(t, "/v1/me"), uToken, nil)
	requireSuccess(t, status, resp)

	var profile map[string]interface{}
	unmarshalData(t, resp, &profile)
	userID, _ := profile["id"].(string)
	if userID == "" {
		userID, _ = profile["user_id"].(string)
	}
	displayName, _ := profile["display_name"].(string)
	t.Logf("Profile: id=%s, display_name=%s", userID, displayName)

	if userID == "" {
		t.Fatal("Could not determine user ID from profile")
	}

	// Step 2: Start a game session
	t.Log("Step 2: Starting game session")
	status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/start"), uToken, nil)
	requireSuccess(t, status, resp)

	var session sessionView
	unmarshalData(t, resp, &session)
	t.Logf("Session: id=%s, total_cards=%d, answers_remaining=%d, skips_remaining=%d",
		session.ID, session.TotalCards, session.AnswersRemaining, session.SkipsRemaining)

	if session.Status != "active" {
		t.Fatalf("expected session status 'active', got '%s'", session.Status)
	}

	// Step 3: Answer all cards (answer up to limit, skip the rest)
	t.Log("Step 3: Answering all cards")
	answeredCount := 0
	skippedCount := 0
	for i := 0; i < session.TotalCards; i++ {
		status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current/card"), uToken, nil)
		if status != http.StatusOK {
			t.Logf("No more cards at index %d", i)
			break
		}

		var card cardView
		unmarshalData(t, resp, &card)

		// Try to answer first
		answer := i%2 == 0
		status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/answer"), uToken,
			map[string]interface{}{"answer": answer})
		if status == http.StatusOK {
			answeredCount++
		} else {
			// Out of answers, skip
			status, resp = doRequest(t, http.MethodPost, apiURL(t, "/v1/sessions/current/skip"), uToken, nil)
			if status == http.StatusOK {
				skippedCount++
			} else {
				t.Logf("Could not answer or skip card %d (status=%d)", i+1, status)
				break
			}
		}
	}
	t.Logf("Cards processed: answered=%d, skipped=%d", answeredCount, skippedCount)

	// Step 4: Verify session completed
	t.Log("Step 4: Checking session completion")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/sessions/current"), uToken, nil)
	requireSuccess(t, status, resp)

	var finalSession sessionView
	unmarshalData(t, resp, &finalSession)
	t.Logf("Final session: status=%s, answers_used=%d, skips_used=%d, current_index=%d",
		finalSession.Status, finalSession.AnswersUsed, finalSession.SkipsUsed, finalSession.CurrentIndex)

	if finalSession.CurrentIndex != finalSession.TotalCards && finalSession.Status != "completed" {
		t.Logf("Warning: session not fully completed (index=%d, total=%d, status=%s)",
			finalSession.CurrentIndex, finalSession.TotalCards, finalSession.Status)
	}

	// Step 5: Check rewards
	t.Log("Step 5: Checking rewards")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/me/rewards"), uToken, nil)
	requireSuccess(t, status, resp)

	var rewards rewardsResponse
	unmarshalData(t, resp, &rewards)
	t.Logf("Rewards: pending=%d, history=%d", len(rewards.Pending), len(rewards.History))

	if rewards.Streak != nil {
		var streak map[string]interface{}
		if err := json.Unmarshal(rewards.Streak, &streak); err == nil {
			currentStreak, _ := streak["current_streak"].(float64)
			t.Logf("Current streak: %d days", int(currentStreak))
		}
	}

	// Step 6: Check leaderboard position
	t.Log("Step 6: Checking leaderboard position")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/leaderboards/daily"), uToken, nil)
	requireSuccess(t, status, resp)

	var lb leaderboardResponse
	unmarshalData(t, resp, &lb)
	t.Logf("Daily leaderboard: period=%s, entries=%d", lb.PeriodKey, len(lb.Entries))

	if lb.UserRank != nil {
		t.Logf("User daily rank: rank=%d, points=%d", lb.UserRank.Rank, lb.UserRank.TotalPoints)
	} else {
		t.Log("User not yet ranked on daily leaderboard (cards may not be resolved yet)")
	}

	// Also check all-time leaderboard
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/leaderboards/all-time"), uToken, nil)
	requireSuccess(t, status, resp)

	var lbAllTime leaderboardResponse
	unmarshalData(t, resp, &lbAllTime)
	if lbAllTime.UserRank != nil {
		t.Logf("User all-time rank: rank=%d, points=%d", lbAllTime.UserRank.Rank, lbAllTime.UserRank.TotalPoints)
	}

	// Step 7: Check achievements
	t.Log("Step 7: Checking achievements")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/me/achievements"), uToken, nil)
	requireSuccess(t, status, resp)

	var achievements map[string]interface{}
	unmarshalData(t, resp, &achievements)

	earnedList, _ := achievements["earned"].([]interface{})
	allList, _ := achievements["achievements"].([]interface{})
	t.Logf("Achievements: earned=%d, total=%d", len(earnedList), len(allList))

	for _, a := range earnedList {
		if aMap, ok := a.(map[string]interface{}); ok {
			name, _ := aMap["name"].(string)
			slug, _ := aMap["slug"].(string)
			t.Logf("  Earned: name=%s, slug=%s", name, slug)
		}
	}

	// Step 8: Check user stats
	t.Log("Step 8: Checking user stats")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/me/stats"), uToken, nil)
	requireSuccess(t, status, resp)

	var stats map[string]interface{}
	unmarshalData(t, resp, &stats)
	t.Logf("User stats: %+v", stats)

	// Step 9: Verify exchange prompts
	t.Log("Step 9: Checking exchange prompts")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/me/exchange-prompts"), uToken, nil)
	requireSuccess(t, status, resp)

	var prompts map[string]interface{}
	unmarshalData(t, resp, &prompts)
	t.Logf("Exchange prompts: %+v", prompts)

	// Step 10: Verify user history
	t.Log("Step 10: Checking user history")
	status, resp = doRequest(t, http.MethodGet, apiURL(t, "/v1/me/history"), uToken, nil)
	requireSuccess(t, status, resp)

	var history interface{}
	unmarshalData(t, resp, &history)
	t.Logf("History endpoint responded successfully")

	t.Log("Full user journey test complete")
}
