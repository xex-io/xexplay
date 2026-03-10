package domain

import (
	"time"

	"github.com/google/uuid"
)

// MiniLeague represents a user-created mini league.
type MiniLeague struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	CreatorID   uuid.UUID  `json:"creator_id"`
	InviteCode  string     `json:"invite_code"`
	EventID     *uuid.UUID `json:"event_id,omitempty"`
	MaxMembers  int        `json:"max_members"`
	MemberCount int        `json:"member_count,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// MiniLeagueMember represents a member of a mini league.
type MiniLeagueMember struct {
	ID       uuid.UUID `json:"id"`
	LeagueID uuid.UUID `json:"league_id"`
	UserID   uuid.UUID `json:"user_id"`
	JoinedAt time.Time `json:"joined_at"`
}
