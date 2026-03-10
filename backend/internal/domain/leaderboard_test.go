package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPeriodTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"PeriodDaily", PeriodDaily, "daily"},
		{"PeriodWeekly", PeriodWeekly, "weekly"},
		{"PeriodTournament", PeriodTournament, "tournament"},
		{"PeriodAllTime", PeriodAllTime, "all_time"},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.expected)
		}
	}
}

func TestLeaderboardEntryCreation(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	eventID := uuid.New()
	now := time.Now()

	entry := LeaderboardEntry{
		ID:             id,
		UserID:         userID,
		EventID:        &eventID,
		PeriodType:     PeriodDaily,
		PeriodKey:      "2026-03-10",
		TotalPoints:    150,
		CorrectAnswers: 8,
		WrongAnswers:   2,
		TotalAnswers:   10,
		Rank:           1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if entry.ID != id {
		t.Errorf("ID mismatch")
	}
	if entry.UserID != userID {
		t.Errorf("UserID mismatch")
	}
	if entry.EventID == nil || *entry.EventID != eventID {
		t.Errorf("EventID mismatch")
	}
	if entry.TotalPoints != 150 {
		t.Errorf("TotalPoints = %d, want 150", entry.TotalPoints)
	}
	if entry.CorrectAnswers != 8 {
		t.Errorf("CorrectAnswers = %d, want 8", entry.CorrectAnswers)
	}
	if entry.Rank != 1 {
		t.Errorf("Rank = %d, want 1", entry.Rank)
	}
}

func TestLeaderboardRowJSONTags(t *testing.T) {
	row := LeaderboardRow{
		Rank:           1,
		UserID:         uuid.New(),
		DisplayName:    "TestUser",
		AvatarURL:      "https://example.com/avatar.png",
		TotalPoints:    200,
		CorrectAnswers: 15,
	}

	data, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("failed to marshal LeaderboardRow: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"rank", "user_id", "display_name", "avatar_url", "total_points", "correct_answers"}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("JSON output missing key %q", key)
		}
	}
}

func TestLeaderboardResponseJSONTags(t *testing.T) {
	resp := LeaderboardResponse{
		PeriodType: PeriodWeekly,
		PeriodKey:  "2026-W10",
		Entries:    []LeaderboardRow{},
		Total:      0,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal LeaderboardResponse: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"period_type", "period_key", "entries", "total"}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("JSON output missing key %q", key)
		}
	}

	// user_rank should be omitted when nil
	if _, ok := m["user_rank"]; ok {
		t.Error("user_rank should be omitted when nil")
	}
}
