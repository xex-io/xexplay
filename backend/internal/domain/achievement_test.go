package domain

import (
	"testing"
)

func TestConditionTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"ConditionFirstPrediction", ConditionFirstPrediction, "first_prediction"},
		{"ConditionPerfectDay", ConditionPerfectDay, "perfect_day"},
		{"ConditionStreak10", ConditionStreak10, "streak_10"},
		{"ConditionStreak30", ConditionStreak30, "streak_30"},
		{"ConditionChampion", ConditionChampion, "champion"},
		{"ConditionReferrals5", ConditionReferrals5, "referrals_5"},
		{"ConditionReferrals10", ConditionReferrals10, "referrals_10"},
		{"ConditionSessions50", ConditionSessions50, "sessions_50"},
		{"ConditionSessions100", ConditionSessions100, "sessions_100"},
		{"ConditionCorrect500", ConditionCorrect500, "correct_500"},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.expected)
		}
	}
}

func TestAllExpectedConditionTypesExist(t *testing.T) {
	// Collect all condition type constants into a set
	allConditions := map[string]bool{
		ConditionFirstPrediction: true,
		ConditionPerfectDay:      true,
		ConditionStreak10:        true,
		ConditionStreak30:        true,
		ConditionChampion:        true,
		ConditionReferrals5:      true,
		ConditionReferrals10:     true,
		ConditionSessions50:      true,
		ConditionSessions100:     true,
		ConditionCorrect500:      true,
	}

	expectedConditions := []string{
		"first_prediction",
		"perfect_day",
		"streak_10",
		"streak_30",
		"champion",
		"referrals_5",
		"referrals_10",
		"sessions_50",
		"sessions_100",
		"correct_500",
	}

	if len(allConditions) != len(expectedConditions) {
		t.Errorf("expected %d condition types, got %d unique constants", len(expectedConditions), len(allConditions))
	}

	for _, cond := range expectedConditions {
		if !allConditions[cond] {
			t.Errorf("missing condition type %q", cond)
		}
	}
}

func TestConditionTypesAreDistinct(t *testing.T) {
	conditions := []string{
		ConditionFirstPrediction,
		ConditionPerfectDay,
		ConditionStreak10,
		ConditionStreak30,
		ConditionChampion,
		ConditionReferrals5,
		ConditionReferrals10,
		ConditionSessions50,
		ConditionSessions100,
		ConditionCorrect500,
	}

	seen := make(map[string]bool)
	for _, c := range conditions {
		if seen[c] {
			t.Errorf("duplicate condition type: %q", c)
		}
		seen[c] = true
	}
}
