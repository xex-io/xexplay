package service

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/rs/zerolog/log"
	"github.com/xex-exchange/xexplay-api/internal/domain"
	"google.golang.org/api/option"
)

// FCMSender implements NotificationSender using Firebase Cloud Messaging.
type FCMSender struct {
	client *messaging.Client
}

// NewFCMSender creates a new FCMSender with the given service account credentials file.
func NewFCMSender(credentialsFile string) (*FCMSender, error) {
	ctx := context.Background()
	opt := option.WithCredentialsFile(credentialsFile)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("init firebase app: %w", err)
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("init firebase messaging: %w", err)
	}
	return &FCMSender{client: client}, nil
}

// NewFCMSenderFromJSON creates a new FCMSender from a JSON credentials string.
func NewFCMSenderFromJSON(credentialsJSON string) (*FCMSender, error) {
	ctx := context.Background()
	opt := option.WithCredentialsJSON([]byte(credentialsJSON))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("init firebase app: %w", err)
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("init firebase messaging: %w", err)
	}
	return &FCMSender{client: client}, nil
}

// Send sends a push notification to a single device token via FCM.
func (s *FCMSender) Send(ctx context.Context, token string, notification *domain.Notification) error {
	msg := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: notification.Title,
			Body:  notification.Body,
		},
		Data: notification.Data,
	}
	_, err := s.client.Send(ctx, msg)
	if err != nil {
		return fmt.Errorf("fcm send: %w", err)
	}
	return nil
}

// SendBatch sends a push notification to multiple device tokens via FCM.
func (s *FCMSender) SendBatch(ctx context.Context, tokens []string, notification *domain.Notification) error {
	var messages []*messaging.Message
	for _, token := range tokens {
		messages = append(messages, &messaging.Message{
			Token: token,
			Notification: &messaging.Notification{
				Title: notification.Title,
				Body:  notification.Body,
			},
			Data: notification.Data,
		})
	}
	resp, err := s.client.SendEach(ctx, messages)
	if err != nil {
		return fmt.Errorf("fcm send batch: %w", err)
	}
	if resp.FailureCount > 0 {
		log.Warn().
			Int("failures", resp.FailureCount).
			Int("successes", resp.SuccessCount).
			Msg("fcm batch partial failure")
	}
	return nil
}
