//go:build integration

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

// testDB sets up a connection to the integration test database.
// It expects DATABASE_URL to be set (e.g., postgres://user:pass@localhost:5432/xexplay_test).
func testDB(t *testing.T) *postgres.DB {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	db, err := postgres.NewConnection(dbURL)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return db
}

// seedTestData creates a user, event, match, basket with 15 cards (3 gold, 5 silver, 7 white),
// publishes the basket, and returns the IDs needed for the game flow test.
type testSeed struct {
	userID   uuid.UUID
	basketID uuid.UUID
	cardIDs  []uuid.UUID // 15 cards in position order
}

func seedTestData(t *testing.T, db *postgres.DB) *testSeed {
	t.Helper()
	ctx := context.Background()
	pool := db.Pool

	// Create a test user
	userID := uuid.New()
	xexUserID := uuid.New()
	referralCode := fmt.Sprintf("TEST-%s", userID.String()[:8])
	_, err := pool.Exec(ctx, `
		INSERT INTO users (id, xex_user_id, display_name, email, role, referral_code, language, total_points, is_active, created_at, updated_at)
		VALUES ($1, $2, 'Test Player', 'test@example.com', 'user', $3, 'en', 0, true, NOW(), NOW())`,
		userID, xexUserID, referralCode)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}

	// Create an event
	eventID := uuid.New()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	_, err = pool.Exec(ctx, `
		INSERT INTO events (id, name, slug, start_date, end_date, is_active, created_at, updated_at)
		VALUES ($1, '{"en": "Test Cup"}', $2, $3, $4, true, NOW(), NOW())`,
		eventID, fmt.Sprintf("test-cup-%s", eventID.String()[:8]), today, today.Add(30*24*time.Hour))
	if err != nil {
		t.Fatalf("seed event: %v", err)
	}

	// Create a match
	matchID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO matches (id, event_id, home_team, away_team, kickoff_time, status, created_at, updated_at)
		VALUES ($1, $2, 'Team A', 'Team B', $3, 'upcoming', NOW(), NOW())`,
		matchID, eventID, today.Add(18*time.Hour))
	if err != nil {
		t.Fatalf("seed match: %v", err)
	}

	// Create 15 cards: 3 gold, 5 silver, 7 white
	type cardSpec struct {
		tier            string
		highAnswerIsYes *bool
	}

	boolTrue := true
	boolFalse := false

	specs := []cardSpec{
		// 3 gold cards
		{domain.TierGold, &boolTrue},
		{domain.TierGold, &boolFalse},
		{domain.TierGold, &boolTrue},
		// 5 silver cards
		{domain.TierSilver, &boolTrue},
		{domain.TierSilver, &boolFalse},
		{domain.TierSilver, &boolTrue},
		{domain.TierSilver, &boolFalse},
		{domain.TierSilver, &boolTrue},
		// 7 white cards
		{domain.TierWhite, nil},
		{domain.TierWhite, nil},
		{domain.TierWhite, nil},
		{domain.TierWhite, nil},
		{domain.TierWhite, nil},
		{domain.TierWhite, nil},
		{domain.TierWhite, nil},
	}

	cardIDs := make([]uuid.UUID, len(specs))
	questionText, _ := json.Marshal(map[string]string{"en": "Will team A win?", "fa": "آیا تیم A می‌بره؟"})

	for i, spec := range specs {
		cardIDs[i] = uuid.New()
		_, err = pool.Exec(ctx, `
			INSERT INTO cards (id, match_id, question_text, tier, high_answer_is_yes, is_resolved, available_date, expires_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, false, $6, $7, NOW(), NOW())`,
			cardIDs[i], matchID, questionText, spec.tier, spec.highAnswerIsYes, today, today.Add(24*time.Hour))
		if err != nil {
			t.Fatalf("seed card %d (%s): %v", i+1, spec.tier, err)
		}
	}

	// Create a basket for today and add all 15 cards
	basketID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO daily_baskets (id, basket_date, event_id, is_published, created_at)
		VALUES ($1, $2, $3, false, NOW())`,
		basketID, today, eventID)
	if err != nil {
		t.Fatalf("seed basket: %v", err)
	}

	// Add cards to basket in position order
	for i, cardID := range cardIDs {
		_, err = pool.Exec(ctx, `
			INSERT INTO daily_basket_cards (id, basket_id, card_id, position)
			VALUES ($1, $2, $3, $4)`,
			uuid.New(), basketID, cardID, i+1)
		if err != nil {
			t.Fatalf("seed basket card %d: %v", i+1, err)
		}
	}

	// Publish the basket (using repo so tier validation runs)
	basketRepo := postgres.NewBasketRepo(db)
	if err := basketRepo.Publish(ctx, basketID); err != nil {
		t.Fatalf("publish basket: %v", err)
	}

	// Register cleanup to delete test data when the test finishes
	t.Cleanup(func() {
		ctx := context.Background()
		pool.Exec(ctx, `DELETE FROM user_answers WHERE user_id = $1`, userID)
		pool.Exec(ctx, `DELETE FROM user_sessions WHERE user_id = $1`, userID)
		pool.Exec(ctx, `DELETE FROM daily_basket_cards WHERE basket_id = $1`, basketID)
		pool.Exec(ctx, `DELETE FROM daily_baskets WHERE id = $1`, basketID)
		for _, cid := range cardIDs {
			pool.Exec(ctx, `DELETE FROM cards WHERE id = $1`, cid)
		}
		pool.Exec(ctx, `DELETE FROM matches WHERE id = $1`, matchID)
		pool.Exec(ctx, `DELETE FROM events WHERE id = $1`, eventID)
		pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, userID)
	})

	return &testSeed{
		userID:   userID,
		basketID: basketID,
		cardIDs:  cardIDs,
	}
}

// newTestGameService creates a GameService wired to the test database with no Redis or streak/achievement services.
func newTestGameService(db *postgres.DB) *GameService {
	sessionRepo := postgres.NewSessionRepo(db)
	answerRepo := postgres.NewAnswerRepo(db)
	basketRepo := postgres.NewBasketRepo(db)
	cardRepo := postgres.NewCardRepo(db)
	userRepo := postgres.NewUserRepo(db)
	shuffleService := NewShuffleService()

	return NewGameService(
		sessionRepo,
		answerRepo,
		basketRepo,
		cardRepo,
		userRepo,
		nil, // no Redis cache for integration tests
		shuffleService,
		nil, // no streak service
		nil, // no achievement service
	)
}

// TestGameFlow_FullSession tests the complete game flow:
// start session -> answer 10 cards -> skip 5 cards -> verify completed session.
func TestGameFlow_FullSession(t *testing.T) {
	db := testDB(t)
	seed := seedTestData(t, db)
	svc := newTestGameService(db)
	ctx := context.Background()

	// --- Step 1: Start a session ---
	session, err := svc.StartSession(ctx, seed.userID)
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}

	if session.Status != domain.SessionStatusActive {
		t.Errorf("new session status = %q, want %q", session.Status, domain.SessionStatusActive)
	}
	if session.CurrentIndex != 0 {
		t.Errorf("new session current_index = %d, want 0", session.CurrentIndex)
	}
	if session.AnswersUsed != 0 {
		t.Errorf("new session answers_used = %d, want 0", session.AnswersUsed)
	}
	if session.SkipsUsed != 0 {
		t.Errorf("new session skips_used = %d, want 0", session.SkipsUsed)
	}
	if len(session.ShuffleOrder) != domain.TotalCards {
		t.Fatalf("shuffle order length = %d, want %d", len(session.ShuffleOrder), domain.TotalCards)
	}

	// --- Step 2: Resume session returns same session ---
	resumed, err := svc.StartSession(ctx, seed.userID)
	if err != nil {
		t.Fatalf("StartSession (resume): %v", err)
	}
	if resumed.ID != session.ID {
		t.Errorf("resumed session ID = %v, want %v (should be same session)", resumed.ID, session.ID)
	}

	// --- Step 3: Answer 10 cards (mix of yes/no) ---
	answerPattern := []bool{true, false, true, true, false, true, false, false, true, true}
	for i, answer := range answerPattern {
		// Get the current card first (sets CardPresentedAt)
		card, err := svc.GetCurrentCard(ctx, session)
		if err != nil {
			t.Fatalf("GetCurrentCard at index %d: %v", i, err)
		}
		if card == nil {
			t.Fatalf("GetCurrentCard returned nil at index %d", i)
		}

		result, err := svc.SubmitAnswer(ctx, session, seed.userID, answer)
		if err != nil {
			t.Fatalf("SubmitAnswer %d (answer=%v): %v", i+1, answer, err)
		}

		if result.AutoSkipped {
			t.Errorf("answer %d was auto-skipped unexpectedly", i+1)
		}

		expectedAnswers := i + 1
		expectedIndex := i + 1
		if session.AnswersUsed != expectedAnswers {
			t.Errorf("after answer %d: answers_used = %d, want %d", i+1, session.AnswersUsed, expectedAnswers)
		}
		if session.CurrentIndex != expectedIndex {
			t.Errorf("after answer %d: current_index = %d, want %d", i+1, session.CurrentIndex, expectedIndex)
		}
	}

	// Verify mid-session state: 10 answers used, 0 skips used, 5 cards remaining
	if session.AnswersUsed != domain.MaxAnswers {
		t.Errorf("after all answers: answers_used = %d, want %d", session.AnswersUsed, domain.MaxAnswers)
	}
	if session.AnswersRemaining() != 0 {
		t.Errorf("after all answers: answers_remaining = %d, want 0", session.AnswersRemaining())
	}
	if session.SkipsUsed != 0 {
		t.Errorf("after all answers: skips_used = %d, want 0", session.SkipsUsed)
	}
	if session.CardsRemaining() != 5 {
		t.Errorf("after all answers: cards_remaining = %d, want 5", session.CardsRemaining())
	}
	if session.Status != domain.SessionStatusActive {
		t.Errorf("after all answers: status = %q, want %q", session.Status, domain.SessionStatusActive)
	}

	// --- Step 4: Skip the remaining 5 cards ---
	for i := 0; i < domain.MaxSkips; i++ {
		// Get the current card first (sets CardPresentedAt)
		card, err := svc.GetCurrentCard(ctx, session)
		if err != nil {
			t.Fatalf("GetCurrentCard for skip %d: %v", i+1, err)
		}
		if card == nil {
			t.Fatalf("GetCurrentCard returned nil for skip %d", i+1)
		}

		result, err := svc.SkipCard(ctx, session, seed.userID)
		if err != nil {
			t.Fatalf("SkipCard %d: %v", i+1, err)
		}

		if result.AutoSkipped {
			t.Errorf("skip %d was auto-skipped unexpectedly", i+1)
		}

		expectedSkips := i + 1
		expectedIndex := domain.MaxAnswers + i + 1
		if session.SkipsUsed != expectedSkips {
			t.Errorf("after skip %d: skips_used = %d, want %d", i+1, session.SkipsUsed, expectedSkips)
		}
		if session.CurrentIndex != expectedIndex {
			t.Errorf("after skip %d: current_index = %d, want %d", i+1, session.CurrentIndex, expectedIndex)
		}
	}

	// --- Step 5: Verify session is completed ---
	if session.Status != domain.SessionStatusCompleted {
		t.Errorf("final status = %q, want %q", session.Status, domain.SessionStatusCompleted)
	}
	if session.CompletedAt == nil {
		t.Error("completed_at is nil, expected a timestamp")
	}
	if session.CurrentIndex != domain.TotalCards {
		t.Errorf("final current_index = %d, want %d", session.CurrentIndex, domain.TotalCards)
	}
	if session.AnswersUsed != domain.MaxAnswers {
		t.Errorf("final answers_used = %d, want %d", session.AnswersUsed, domain.MaxAnswers)
	}
	if session.SkipsUsed != domain.MaxSkips {
		t.Errorf("final skips_used = %d, want %d", session.SkipsUsed, domain.MaxSkips)
	}
	if !session.IsComplete() {
		t.Error("IsComplete() = false, want true")
	}
	if session.CardsRemaining() != 0 {
		t.Errorf("final cards_remaining = %d, want 0", session.CardsRemaining())
	}

	// --- Step 6: Verify answers were persisted ---
	answerRepo := postgres.NewAnswerRepo(db)
	answers, err := answerRepo.FindBySession(ctx, session.ID)
	if err != nil {
		t.Fatalf("FindBySession: %v", err)
	}
	if len(answers) != domain.MaxAnswers {
		t.Errorf("persisted answers = %d, want %d", len(answers), domain.MaxAnswers)
	}

	// Verify answer values match what we submitted
	for i, a := range answers {
		if a.Answer != answerPattern[i] {
			t.Errorf("answer %d: answer = %v, want %v", i+1, a.Answer, answerPattern[i])
		}
		if a.SessionID != session.ID {
			t.Errorf("answer %d: session_id = %v, want %v", i+1, a.SessionID, session.ID)
		}
		if a.UserID != seed.userID {
			t.Errorf("answer %d: user_id = %v, want %v", i+1, a.UserID, seed.userID)
		}
	}
}

// TestGameFlow_SessionView verifies the SessionView output at key points of the game flow.
func TestGameFlow_SessionView(t *testing.T) {
	db := testDB(t)
	seed := seedTestData(t, db)
	svc := newTestGameService(db)
	ctx := context.Background()

	session, err := svc.StartSession(ctx, seed.userID)
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}

	// Verify initial SessionView
	view := session.ToView()
	if view.TotalCards != domain.TotalCards {
		t.Errorf("view.TotalCards = %d, want %d", view.TotalCards, domain.TotalCards)
	}
	if view.AnswersRemaining != domain.MaxAnswers {
		t.Errorf("view.AnswersRemaining = %d, want %d", view.AnswersRemaining, domain.MaxAnswers)
	}
	if view.SkipsRemaining != domain.MaxSkips {
		t.Errorf("view.SkipsRemaining = %d, want %d", view.SkipsRemaining, domain.MaxSkips)
	}

	// Answer 5 cards, then check the view
	for i := 0; i < 5; i++ {
		_, _ = svc.GetCurrentCard(ctx, session)
		_, err := svc.SubmitAnswer(ctx, session, seed.userID, true)
		if err != nil {
			t.Fatalf("SubmitAnswer %d: %v", i+1, err)
		}
	}

	view = session.ToView()
	if view.AnswersUsed != 5 {
		t.Errorf("mid-view.AnswersUsed = %d, want 5", view.AnswersUsed)
	}
	if view.AnswersRemaining != domain.MaxAnswers-5 {
		t.Errorf("mid-view.AnswersRemaining = %d, want %d", view.AnswersRemaining, domain.MaxAnswers-5)
	}
	if view.CurrentIndex != 5 {
		t.Errorf("mid-view.CurrentIndex = %d, want 5", view.CurrentIndex)
	}
	if view.Status != domain.SessionStatusActive {
		t.Errorf("mid-view.Status = %q, want %q", view.Status, domain.SessionStatusActive)
	}
}

// TestGameFlow_ErrorCases uses table-driven tests to verify error conditions in the game flow.
func TestGameFlow_ErrorCases(t *testing.T) {
	db := testDB(t)
	seed := seedTestData(t, db)
	svc := newTestGameService(db)
	ctx := context.Background()

	session, err := svc.StartSession(ctx, seed.userID)
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}

	wrongUserID := uuid.New()

	tests := []struct {
		name    string
		action  func() error
		wantErr string
	}{
		{
			name: "submit answer with wrong user",
			action: func() error {
				_, _ = svc.GetCurrentCard(ctx, session)
				_, err := svc.SubmitAnswer(ctx, session, wrongUserID, true)
				return err
			},
			wantErr: "session does not belong to user",
		},
		{
			name: "skip card with wrong user",
			action: func() error {
				_, err := svc.SkipCard(ctx, session, wrongUserID)
				return err
			},
			wantErr: "session does not belong to user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.action()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}

// TestGameFlow_ResourceExhaustion verifies that exceeding answer or skip limits returns errors.
func TestGameFlow_ResourceExhaustion(t *testing.T) {
	db := testDB(t)
	seed := seedTestData(t, db)
	svc := newTestGameService(db)
	ctx := context.Background()

	session, err := svc.StartSession(ctx, seed.userID)
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}

	// Use all 10 answers
	for i := 0; i < domain.MaxAnswers; i++ {
		_, _ = svc.GetCurrentCard(ctx, session)
		_, err := svc.SubmitAnswer(ctx, session, seed.userID, true)
		if err != nil {
			t.Fatalf("SubmitAnswer %d: %v", i+1, err)
		}
	}

	// 11th answer should fail with "no answers remaining"
	_, _ = svc.GetCurrentCard(ctx, session)
	_, err = svc.SubmitAnswer(ctx, session, seed.userID, true)
	if err == nil {
		t.Fatal("expected error when answers exhausted, got nil")
	}
	if err.Error() != "no answers remaining" {
		t.Errorf("error = %q, want %q", err.Error(), "no answers remaining")
	}

	// Use all 5 skips
	for i := 0; i < domain.MaxSkips; i++ {
		_, _ = svc.GetCurrentCard(ctx, session)
		_, err := svc.SkipCard(ctx, session, seed.userID)
		if err != nil {
			t.Fatalf("SkipCard %d: %v", i+1, err)
		}
	}

	// Session should be completed now (10 answers + 5 skips = 15 total cards)
	if session.Status != domain.SessionStatusCompleted {
		t.Errorf("status after exhaustion = %q, want %q", session.Status, domain.SessionStatusCompleted)
	}
}

// TestGameFlow_GetCurrentCard_CompletedSession verifies that GetCurrentCard fails on a completed session.
func TestGameFlow_GetCurrentCard_CompletedSession(t *testing.T) {
	db := testDB(t)
	seed := seedTestData(t, db)
	svc := newTestGameService(db)
	ctx := context.Background()

	session, err := svc.StartSession(ctx, seed.userID)
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}

	// Complete the session: 10 answers + 5 skips
	for i := 0; i < domain.MaxAnswers; i++ {
		_, _ = svc.GetCurrentCard(ctx, session)
		_, _ = svc.SubmitAnswer(ctx, session, seed.userID, true)
	}
	for i := 0; i < domain.MaxSkips; i++ {
		_, _ = svc.GetCurrentCard(ctx, session)
		_, _ = svc.SkipCard(ctx, session, seed.userID)
	}

	// GetCurrentCard should fail on completed session
	_, err = svc.GetCurrentCard(ctx, session)
	if err == nil {
		t.Fatal("expected error on completed session, got nil")
	}
	if err.Error() != "session is complete, no more cards" {
		t.Errorf("error = %q, want %q", err.Error(), "session is complete, no more cards")
	}
}

// TestGameFlow_CardTierDistribution verifies that the basket has the correct tier distribution
// when cards are fetched during the session.
func TestGameFlow_CardTierDistribution(t *testing.T) {
	db := testDB(t)
	seed := seedTestData(t, db)
	svc := newTestGameService(db)
	ctx := context.Background()

	session, err := svc.StartSession(ctx, seed.userID)
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}

	// Iterate through all 15 cards and count tiers
	tierCounts := map[string]int{
		domain.TierGold:   0,
		domain.TierSilver: 0,
		domain.TierWhite:  0,
	}

	for i := 0; i < domain.TotalCards; i++ {
		card, err := svc.GetCurrentCard(ctx, session)
		if err != nil {
			t.Fatalf("GetCurrentCard %d: %v", i+1, err)
		}

		tierCounts[card.Tier]++

		// Advance the session (alternate answers and skips)
		if i < domain.MaxAnswers {
			_, err = svc.SubmitAnswer(ctx, session, seed.userID, true)
		} else {
			_, err = svc.SkipCard(ctx, session, seed.userID)
		}
		if err != nil {
			t.Fatalf("advance card %d: %v", i+1, err)
		}
	}

	// Verify tier distribution
	tests := []struct {
		tier string
		want int
	}{
		{domain.TierGold, domain.GoldCount},
		{domain.TierSilver, domain.SilverCount},
		{domain.TierWhite, domain.WhiteCount},
	}

	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			if tierCounts[tt.tier] != tt.want {
				t.Errorf("%s cards = %d, want %d", tt.tier, tierCounts[tt.tier], tt.want)
			}
		})
	}
}
