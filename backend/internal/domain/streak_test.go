package domain

import (
	"testing"
)

func TestCheckMilestone_ReturnsCorrectMilestone(t *testing.T) {
	milestoneDays := []int{3, 7, 10, 14, 21, 30}

	for _, day := range milestoneDays {
		m := CheckMilestone(day)
		if m == nil {
			t.Errorf("CheckMilestone(%d) returned nil, expected a milestone", day)
			continue
		}
		if m.Days != day {
			t.Errorf("CheckMilestone(%d).Days = %d, want %d", day, m.Days, day)
		}
	}
}

func TestCheckMilestone_ReturnsNilForNonMilestoneDays(t *testing.T) {
	nonMilestoneDays := []int{0, 1, 2, 4, 5, 6, 8, 9, 11, 12, 13, 15, 20, 25, 29, 31, 100}

	for _, day := range nonMilestoneDays {
		m := CheckMilestone(day)
		if m != nil {
			t.Errorf("CheckMilestone(%d) returned milestone with Days=%d, expected nil", day, m.Days)
		}
	}
}

func TestGetMilestones_ReturnsSixMilestones(t *testing.T) {
	milestones := GetMilestones()
	if len(milestones) != 6 {
		t.Errorf("GetMilestones() returned %d milestones, want 6", len(milestones))
	}
}

func TestGetMilestones_DaysAreInOrder(t *testing.T) {
	milestones := GetMilestones()
	expectedDays := []int{3, 7, 10, 14, 21, 30}

	for i, m := range milestones {
		if m.Days != expectedDays[i] {
			t.Errorf("GetMilestones()[%d].Days = %d, want %d", i, m.Days, expectedDays[i])
		}
	}
}

func TestMilestoneBonusValues(t *testing.T) {
	tests := []struct {
		days         int
		bonusSkips   int
		bonusAnswers int
		tokenReward  float64
	}{
		{3, 0, 0, 0},
		{7, 1, 0, 0},
		{10, 1, 0, 1.0},
		{14, 0, 1, 0},
		{21, 1, 1, 2.0},
		{30, 2, 1, 5.0},
	}

	for _, tt := range tests {
		m := CheckMilestone(tt.days)
		if m == nil {
			t.Fatalf("CheckMilestone(%d) returned nil", tt.days)
		}
		if m.BonusSkips != tt.bonusSkips {
			t.Errorf("Day %d: BonusSkips = %d, want %d", tt.days, m.BonusSkips, tt.bonusSkips)
		}
		if m.BonusAnswers != tt.bonusAnswers {
			t.Errorf("Day %d: BonusAnswers = %d, want %d", tt.days, m.BonusAnswers, tt.bonusAnswers)
		}
		if m.TokenReward != tt.tokenReward {
			t.Errorf("Day %d: TokenReward = %f, want %f", tt.days, m.TokenReward, tt.tokenReward)
		}
	}
}

func TestMilestoneDescriptionsNotEmpty(t *testing.T) {
	for _, m := range GetMilestones() {
		if m.Description == "" {
			t.Errorf("Milestone at day %d has empty description", m.Days)
		}
	}
}
