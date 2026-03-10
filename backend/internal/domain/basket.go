package domain

import (
	"time"

	"github.com/google/uuid"
)

type DailyBasket struct {
	ID          uuid.UUID `json:"id"`
	BasketDate  time.Time `json:"basket_date"`
	EventID     uuid.UUID `json:"event_id"`
	IsPublished bool      `json:"is_published"`
	CreatedAt   time.Time `json:"created_at"`
}

type DailyBasketCard struct {
	ID       uuid.UUID `json:"id"`
	BasketID uuid.UUID `json:"basket_id"`
	CardID   uuid.UUID `json:"card_id"`
	Position int       `json:"position"` // 1-15
}
