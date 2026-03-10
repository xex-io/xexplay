package domain

import "testing"

func boolPtr(b bool) *bool { return &b }

func TestPointsForAnswer_Gold(t *testing.T) {
	yes := true

	card := &Card{
		Tier:            TierGold,
		HighAnswerIsYes: boolPtr(yes),
	}

	// Gold correct (answer matches HighAnswerIsYes=true, so answer=true) → 20
	if got := card.PointsForAnswer(true); got != GoldHighPoints {
		t.Errorf("Gold correct: got %d, want %d", got, GoldHighPoints)
	}

	// Gold incorrect (answer=false, HighAnswerIsYes=true) → 5
	if got := card.PointsForAnswer(false); got != GoldLowPoints {
		t.Errorf("Gold incorrect: got %d, want %d", got, GoldLowPoints)
	}
}

func TestPointsForAnswer_GoldNilHighAnswer(t *testing.T) {
	card := &Card{
		Tier:            TierGold,
		HighAnswerIsYes: nil,
	}

	if got := card.PointsForAnswer(true); got != GoldLowPoints {
		t.Errorf("Gold nil high answer (yes): got %d, want %d", got, GoldLowPoints)
	}
	if got := card.PointsForAnswer(false); got != GoldLowPoints {
		t.Errorf("Gold nil high answer (no): got %d, want %d", got, GoldLowPoints)
	}
}

func TestPointsForAnswer_Silver(t *testing.T) {
	yes := true

	card := &Card{
		Tier:            TierSilver,
		HighAnswerIsYes: boolPtr(yes),
	}

	// Silver correct → 15
	if got := card.PointsForAnswer(true); got != SilverHighPoints {
		t.Errorf("Silver correct: got %d, want %d", got, SilverHighPoints)
	}

	// Silver incorrect → 10
	if got := card.PointsForAnswer(false); got != SilverLowPoints {
		t.Errorf("Silver incorrect: got %d, want %d", got, SilverLowPoints)
	}
}

func TestPointsForAnswer_White(t *testing.T) {
	card := &Card{
		Tier: TierWhite,
	}

	// White always → 10 regardless of answer
	if got := card.PointsForAnswer(true); got != WhitePoints {
		t.Errorf("White yes: got %d, want %d", got, WhitePoints)
	}
	if got := card.PointsForAnswer(false); got != WhitePoints {
		t.Errorf("White no: got %d, want %d", got, WhitePoints)
	}
}
