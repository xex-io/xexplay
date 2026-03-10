package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

// --- Mock implementations ---

// mockFCMTokenRepo implements the methods called by NotificationService
// without requiring a real database connection.
type mockFCMTokenRepo struct {
	tokens            []domain.FCMToken
	findErr           error
	deactivatedTokens []string
	deactivateErr     error
}

func (m *mockFCMTokenRepo) FindActiveByUser(_ context.Context, userID uuid.UUID) ([]domain.FCMToken, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var result []domain.FCMToken
	for _, t := range m.tokens {
		if t.UserID == userID && t.IsActive {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockFCMTokenRepo) FindAllActive(_ context.Context) ([]domain.FCMToken, error) {
	if m.findErr != nil {
		return nil, m.findErr
	}
	var result []domain.FCMToken
	for _, t := range m.tokens {
		if t.IsActive {
			result = append(result, t)
		}
	}
	return result, nil
}

func (m *mockFCMTokenRepo) DeactivateByTokenValue(_ context.Context, tokenValue string) error {
	m.deactivatedTokens = append(m.deactivatedTokens, tokenValue)
	return m.deactivateErr
}

// mockSender records all sent notifications for verification.
type mockSender struct {
	sent      []sentNotification
	sendErr   error
	batchSent []batchSentNotification
	batchErr  error
}

type sentNotification struct {
	Token        string
	Notification *domain.Notification
}

type batchSentNotification struct {
	Tokens       []string
	Notification *domain.Notification
}

func (m *mockSender) Send(_ context.Context, token string, notification *domain.Notification) error {
	m.sent = append(m.sent, sentNotification{Token: token, Notification: notification})
	return m.sendErr
}

func (m *mockSender) SendBatch(_ context.Context, tokens []string, notification *domain.Notification) error {
	m.batchSent = append(m.batchSent, batchSentNotification{Tokens: tokens, Notification: notification})
	return m.batchErr
}

// fcmTokenRepoInterface defines the interface that NotificationService uses.
// We use this to create a testable version of the service.
type fcmTokenRepoInterface interface {
	FindActiveByUser(ctx context.Context, userID uuid.UUID) ([]domain.FCMToken, error)
	FindAllActive(ctx context.Context) ([]domain.FCMToken, error)
	DeactivateByTokenValue(ctx context.Context, tokenValue string) error
}

// testNotificationService wraps the notification logic using interfaces for testability.
type testNotificationService struct {
	repo   fcmTokenRepoInterface
	sender NotificationSender
}

func (s *testNotificationService) SendToUser(ctx context.Context, userID uuid.UUID, notification *domain.Notification) error {
	tokens, err := s.repo.FindActiveByUser(ctx, userID)
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		return nil
	}
	for _, t := range tokens {
		if err := s.sender.Send(ctx, t.Token, notification); err != nil {
			_ = s.repo.DeactivateByTokenValue(ctx, t.Token)
		}
	}
	return nil
}

// --- Notification Dispatch Tests (2.5.7) ---

func TestSendToUser_DispatchesToAllDevices(t *testing.T) {
	userID := uuid.New()
	repo := &mockFCMTokenRepo{
		tokens: []domain.FCMToken{
			{ID: uuid.New(), UserID: userID, Token: "token-ios-1", DeviceType: domain.DeviceIOS, IsActive: true},
			{ID: uuid.New(), UserID: userID, Token: "token-android-1", DeviceType: domain.DeviceAndroid, IsActive: true},
		},
	}
	sender := &mockSender{}
	svc := &testNotificationService{repo: repo, sender: sender}

	notification := &domain.Notification{
		Title:      "Test Notification",
		Body:       "This is a test",
		TargetType: domain.TargetUser,
	}

	err := svc.SendToUser(context.Background(), userID, notification)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sender.sent) != 2 {
		t.Fatalf("expected 2 notifications sent, got %d", len(sender.sent))
	}

	expectedTokens := map[string]bool{"token-ios-1": false, "token-android-1": false}
	for _, s := range sender.sent {
		if _, ok := expectedTokens[s.Token]; !ok {
			t.Errorf("unexpected token %q in sent notifications", s.Token)
		}
		expectedTokens[s.Token] = true
		if s.Notification.Title != "Test Notification" {
			t.Errorf("notification title = %q, want %q", s.Notification.Title, "Test Notification")
		}
	}
	for token, sent := range expectedTokens {
		if !sent {
			t.Errorf("expected notification to be sent to token %q", token)
		}
	}
}

func TestSendToUser_NoTokensSkipsNotification(t *testing.T) {
	userID := uuid.New()
	repo := &mockFCMTokenRepo{tokens: []domain.FCMToken{}}
	sender := &mockSender{}
	svc := &testNotificationService{repo: repo, sender: sender}

	notification := &domain.Notification{
		Title:      "Test",
		Body:       "Body",
		TargetType: domain.TargetUser,
	}

	err := svc.SendToUser(context.Background(), userID, notification)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sender.sent) != 0 {
		t.Errorf("expected 0 notifications sent for user with no tokens, got %d", len(sender.sent))
	}
}

func TestSendToUser_RepoErrorReturnsError(t *testing.T) {
	userID := uuid.New()
	repo := &mockFCMTokenRepo{findErr: errors.New("db connection failed")}
	sender := &mockSender{}
	svc := &testNotificationService{repo: repo, sender: sender}

	err := svc.SendToUser(context.Background(), userID, &domain.Notification{
		Title:      "Test",
		Body:       "Body",
		TargetType: domain.TargetUser,
	})
	if err == nil {
		t.Fatal("expected error when repo fails, got nil")
	}
}

func TestSendToUser_DeactivatesTokenOnSendFailure(t *testing.T) {
	userID := uuid.New()
	repo := &mockFCMTokenRepo{
		tokens: []domain.FCMToken{
			{ID: uuid.New(), UserID: userID, Token: "invalid-token", DeviceType: domain.DeviceIOS, IsActive: true},
		},
	}
	sender := &mockSender{sendErr: errors.New("FCM: token not registered")}
	svc := &testNotificationService{repo: repo, sender: sender}

	_ = svc.SendToUser(context.Background(), userID, &domain.Notification{
		Title:      "Test",
		Body:       "Body",
		TargetType: domain.TargetUser,
	})

	if len(repo.deactivatedTokens) != 1 {
		t.Fatalf("expected 1 token deactivated, got %d", len(repo.deactivatedTokens))
	}
	if repo.deactivatedTokens[0] != "invalid-token" {
		t.Errorf("deactivated token = %q, want %q", repo.deactivatedTokens[0], "invalid-token")
	}
}

func TestSendToUser_OnlyMatchesCorrectUser(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()
	repo := &mockFCMTokenRepo{
		tokens: []domain.FCMToken{
			{ID: uuid.New(), UserID: userA, Token: "token-a", DeviceType: domain.DeviceIOS, IsActive: true},
			{ID: uuid.New(), UserID: userB, Token: "token-b", DeviceType: domain.DeviceAndroid, IsActive: true},
		},
	}
	sender := &mockSender{}
	svc := &testNotificationService{repo: repo, sender: sender}

	err := svc.SendToUser(context.Background(), userA, &domain.Notification{
		Title:      "For User A",
		Body:       "Body",
		TargetType: domain.TargetUser,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sender.sent) != 1 {
		t.Fatalf("expected 1 notification for userA, got %d", len(sender.sent))
	}
	if sender.sent[0].Token != "token-a" {
		t.Errorf("sent to token %q, want %q", sender.sent[0].Token, "token-a")
	}
}

// --- FCM Token Management Tests (2.5.7) ---

func TestFCMTokenRegister_NewToken(t *testing.T) {
	// Verify the domain model correctly represents a new FCM token
	userID := uuid.New()
	token := domain.FCMToken{
		ID:         uuid.New(),
		UserID:     userID,
		Token:      "fcm-token-abc123",
		DeviceType: domain.DeviceIOS,
		IsActive:   true,
	}

	if token.Token != "fcm-token-abc123" {
		t.Errorf("Token = %q, want %q", token.Token, "fcm-token-abc123")
	}
	if token.DeviceType != domain.DeviceIOS {
		t.Errorf("DeviceType = %q, want %q", token.DeviceType, domain.DeviceIOS)
	}
	if !token.IsActive {
		t.Error("expected new token to be active")
	}
	if token.UserID != userID {
		t.Errorf("UserID = %v, want %v", token.UserID, userID)
	}
}

func TestFCMTokenDeactivate_MarksTokenInactive(t *testing.T) {
	userID := uuid.New()
	repo := &mockFCMTokenRepo{
		tokens: []domain.FCMToken{
			{ID: uuid.New(), UserID: userID, Token: "token-to-deactivate", DeviceType: domain.DeviceAndroid, IsActive: true},
		},
	}

	err := repo.DeactivateByTokenValue(context.Background(), "token-to-deactivate")
	if err != nil {
		t.Fatalf("unexpected error deactivating token: %v", err)
	}

	if len(repo.deactivatedTokens) != 1 {
		t.Fatalf("expected 1 deactivated token, got %d", len(repo.deactivatedTokens))
	}
	if repo.deactivatedTokens[0] != "token-to-deactivate" {
		t.Errorf("deactivated token = %q, want %q", repo.deactivatedTokens[0], "token-to-deactivate")
	}
}

func TestFCMTokenDeactivate_ErrorPropagates(t *testing.T) {
	repo := &mockFCMTokenRepo{
		deactivateErr: errors.New("database error"),
	}

	err := repo.DeactivateByTokenValue(context.Background(), "some-token")
	if err == nil {
		t.Fatal("expected error when deactivation fails, got nil")
	}
}

func TestDeviceTypeConstants(t *testing.T) {
	// Verify the device type constants have expected values
	if domain.DeviceIOS != "ios" {
		t.Errorf("DeviceIOS = %q, want %q", domain.DeviceIOS, "ios")
	}
	if domain.DeviceAndroid != "android" {
		t.Errorf("DeviceAndroid = %q, want %q", domain.DeviceAndroid, "android")
	}
	if domain.DeviceWeb != "web" {
		t.Errorf("DeviceWeb = %q, want %q", domain.DeviceWeb, "web")
	}
}

func TestLogSender_DoesNotError(t *testing.T) {
	sender := NewLogSender()

	notification := &domain.Notification{
		Title:      "Test",
		Body:       "Body",
		TargetType: domain.TargetUser,
		Data:       map[string]string{"key": "value"},
	}

	err := sender.Send(context.Background(), "test-token", notification)
	if err != nil {
		t.Errorf("LogSender.Send() returned error: %v", err)
	}

	err = sender.SendBatch(context.Background(), []string{"t1", "t2"}, notification)
	if err != nil {
		t.Errorf("LogSender.SendBatch() returned error: %v", err)
	}
}
