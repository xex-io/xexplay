package domain

import (
	"testing"
	"time"
)

func TestAnswersRemaining(t *testing.T) {
	tests := []struct {
		name         string
		answersUsed  int
		bonusAnswers int
		want         int
	}{
		{"fresh session", 0, 0, MaxAnswers},
		{"some used", 3, 0, MaxAnswers - 3},
		{"all used", MaxAnswers, 0, 0},
		{"with bonus", 5, 2, MaxAnswers + 2 - 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UserSession{AnswersUsed: tt.answersUsed, BonusAnswers: tt.bonusAnswers}
			if got := s.AnswersRemaining(); got != tt.want {
				t.Errorf("AnswersRemaining() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSkipsRemaining(t *testing.T) {
	tests := []struct {
		name       string
		skipsUsed  int
		bonusSkips int
		want       int
	}{
		{"fresh session", 0, 0, MaxSkips},
		{"some used", 2, 0, MaxSkips - 2},
		{"all used", MaxSkips, 0, 0},
		{"with bonus", 3, 1, MaxSkips + 1 - 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UserSession{SkipsUsed: tt.skipsUsed, BonusSkips: tt.bonusSkips}
			if got := s.SkipsRemaining(); got != tt.want {
				t.Errorf("SkipsRemaining() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCardsRemaining(t *testing.T) {
	tests := []struct {
		name         string
		currentIndex int
		want         int
	}{
		{"start", 0, TotalCards},
		{"midway", 7, TotalCards - 7},
		{"last card", TotalCards - 1, 1},
		{"done", TotalCards, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UserSession{CurrentIndex: tt.currentIndex}
			if got := s.CardsRemaining(); got != tt.want {
				t.Errorf("CardsRemaining() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIsComplete(t *testing.T) {
	tests := []struct {
		name         string
		currentIndex int
		want         bool
	}{
		{"not complete", 0, false},
		{"midway", 7, false},
		{"exactly total", TotalCards, true},
		{"beyond total", TotalCards + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &UserSession{CurrentIndex: tt.currentIndex}
			if got := s.IsComplete(); got != tt.want {
				t.Errorf("IsComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCardExpired(t *testing.T) {
	t.Run("nil CardPresentedAt returns false", func(t *testing.T) {
		s := &UserSession{CardPresentedAt: nil}
		if s.IsCardExpired() {
			t.Error("expected false when CardPresentedAt is nil")
		}
	})

	t.Run("within timer returns false", func(t *testing.T) {
		now := time.Now()
		s := &UserSession{CardPresentedAt: &now}
		if s.IsCardExpired() {
			t.Error("expected false when card was just presented")
		}
	})

	t.Run("within 42s returns false", func(t *testing.T) {
		presented := time.Now().Add(-41 * time.Second)
		s := &UserSession{CardPresentedAt: &presented}
		if s.IsCardExpired() {
			t.Error("expected false within 42s window")
		}
	})

	t.Run("after 42s returns true", func(t *testing.T) {
		presented := time.Now().Add(-43 * time.Second)
		s := &UserSession{CardPresentedAt: &presented}
		if !s.IsCardExpired() {
			t.Error("expected true after 42s (40s timer + 2s grace)")
		}
	})
}
