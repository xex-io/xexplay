package domain

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestRewardTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"RewardToken", RewardToken, "token"},
		{"RewardBonusSkip", RewardBonusSkip, "bonus_skip"},
		{"RewardBonusAnswer", RewardBonusAnswer, "bonus_answer"},
		{"RewardBadge", RewardBadge, "badge"},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.expected)
		}
	}
}

func TestRewardStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"StatusPending", StatusPending, "pending"},
		{"StatusClaimed", StatusClaimed, "claimed"},
		{"StatusCredited", StatusCredited, "credited"},
		{"StatusExpired", StatusExpired, "expired"},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.expected)
		}
	}
}

func TestRewardConfigStructFields(t *testing.T) {
	id := uuid.New()
	cfg := RewardConfig{
		ID:          id,
		PeriodType:  PeriodDaily,
		RankFrom:    1,
		RankTo:      3,
		RewardType:  RewardToken,
		Amount:      10.5,
		Description: json.RawMessage(`{"en":"Top 3 daily reward"}`),
		IsActive:    true,
	}

	if cfg.ID != id {
		t.Error("ID mismatch")
	}
	if cfg.PeriodType != PeriodDaily {
		t.Errorf("PeriodType = %q, want %q", cfg.PeriodType, PeriodDaily)
	}
	if cfg.RankFrom != 1 || cfg.RankTo != 3 {
		t.Errorf("Rank range = [%d, %d], want [1, 3]", cfg.RankFrom, cfg.RankTo)
	}
	if cfg.RewardType != RewardToken {
		t.Errorf("RewardType = %q, want %q", cfg.RewardType, RewardToken)
	}
	if cfg.Amount != 10.5 {
		t.Errorf("Amount = %f, want 10.5", cfg.Amount)
	}
	if !cfg.IsActive {
		t.Error("IsActive should be true")
	}
}

func TestRewardConfigJSONMarshal(t *testing.T) {
	cfg := RewardConfig{
		ID:         uuid.New(),
		PeriodType: PeriodWeekly,
		RankFrom:   1,
		RankTo:     10,
		RewardType: RewardBadge,
		Amount:     0,
		IsActive:   true,
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal RewardConfig: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	requiredKeys := []string{"id", "period_type", "rank_from", "rank_to", "reward_type", "amount", "is_active"}
	for _, key := range requiredKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("JSON output missing key %q", key)
		}
	}
}

func TestRewardDistributionStatusFlow(t *testing.T) {
	// Verify the expected status progression exists as constants
	statuses := []string{StatusPending, StatusClaimed, StatusCredited, StatusExpired}
	if len(statuses) != 4 {
		t.Errorf("expected 4 status constants, got %d", len(statuses))
	}

	// Verify each status is a distinct value
	seen := make(map[string]bool)
	for _, s := range statuses {
		if seen[s] {
			t.Errorf("duplicate status constant: %q", s)
		}
		seen[s] = true
	}
}
