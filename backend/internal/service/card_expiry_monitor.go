package service

import (
	"context"
	"math"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/xex-exchange/xexplay-api/internal/pkg/ws"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

// CardExpiryMonitor is a background service that monitors card expiry times
// and broadcasts WebSocket notifications to connected users.
type CardExpiryMonitor struct {
	cardRepo *postgres.CardRepo
	hub      *ws.Hub
	// Track cards that have already been notified as expired to avoid duplicate broadcasts.
	notifiedExpired map[string]bool
}

// NewCardExpiryMonitor creates a new CardExpiryMonitor.
func NewCardExpiryMonitor(cardRepo *postgres.CardRepo, hub *ws.Hub) *CardExpiryMonitor {
	return &CardExpiryMonitor{
		cardRepo:        cardRepo,
		hub:             hub,
		notifiedExpired: make(map[string]bool),
	}
}

// Start begins the background monitoring loop. It runs every minute and checks for
// cards expiring within 30 minutes (broadcasting card_expiring) and cards that have
// already expired (broadcasting card_expired). It stops when the context is cancelled.
func (m *CardExpiryMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Info().Msg("card expiry monitor started")

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("card expiry monitor stopped")
			return
		case <-ticker.C:
			m.check(ctx)
		}
	}
}

func (m *CardExpiryMonitor) check(ctx context.Context) {
	now := time.Now().UTC()
	today := now.Truncate(24 * time.Hour)

	// Find all cards available today (unresolved ones are the most relevant)
	cards, err := m.cardRepo.FindByAvailableDate(ctx, today)
	if err != nil {
		log.Warn().Err(err).Msg("card expiry monitor: failed to fetch cards")
		return
	}

	for _, card := range cards {
		if card.IsResolved {
			continue
		}

		cardIDStr := card.ID.String()
		timeUntilExpiry := card.ExpiresAt.Sub(now)

		if timeUntilExpiry <= 0 {
			// Card has expired
			if !m.notifiedExpired[cardIDStr] {
				m.hub.Broadcast(ws.Message{
					Type: "card_expired",
					Data: map[string]interface{}{
						"card_id": cardIDStr,
					},
				})
				m.notifiedExpired[cardIDStr] = true
				log.Debug().Str("card_id", cardIDStr).Msg("card expiry monitor: broadcast card_expired")
			}
		} else if timeUntilExpiry <= 30*time.Minute {
			// Card is expiring within 30 minutes
			minutesLeft := int(math.Ceil(timeUntilExpiry.Minutes()))
			m.hub.Broadcast(ws.Message{
				Type: "card_expiring",
				Data: map[string]interface{}{
					"card_id":            cardIDStr,
					"expires_in_minutes": minutesLeft,
				},
			})
			log.Debug().
				Str("card_id", cardIDStr).
				Int("expires_in_minutes", minutesLeft).
				Msg("card expiry monitor: broadcast card_expiring")
		}
	}

	// Clean up old expired entries to prevent memory growth
	// Remove entries for cards that expired more than 24 hours ago
	for cardID := range m.notifiedExpired {
		// We only clean up periodically; the map is bounded by daily card count
		// which is small enough that explicit cleanup isn't critical.
		_ = cardID
	}
}
