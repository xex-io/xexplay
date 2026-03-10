package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type MiniLeagueRepo struct {
	db *DB
}

func NewMiniLeagueRepo(db *DB) *MiniLeagueRepo {
	return &MiniLeagueRepo{db: db}
}

// Create inserts a new mini league.
func (r *MiniLeagueRepo) Create(ctx context.Context, league *domain.MiniLeague) error {
	query := `
		INSERT INTO mini_leagues (id, name, creator_id, invite_code, event_id, max_members, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING created_at, updated_at`

	err := r.db.Pool.QueryRow(ctx, query,
		league.ID, league.Name, league.CreatorID, league.InviteCode, league.EventID, league.MaxMembers,
	).Scan(&league.CreatedAt, &league.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create mini league: %w", err)
	}
	return nil
}

// FindByID returns a mini league by ID, including a member count.
func (r *MiniLeagueRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.MiniLeague, error) {
	query := `
		SELECT ml.id, ml.name, ml.creator_id, ml.invite_code, ml.event_id, ml.max_members,
		       ml.created_at, ml.updated_at,
		       (SELECT COUNT(*) FROM mini_league_members WHERE league_id = ml.id) AS member_count
		FROM mini_leagues ml
		WHERE ml.id = $1`

	var league domain.MiniLeague
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&league.ID, &league.Name, &league.CreatorID, &league.InviteCode, &league.EventID,
		&league.MaxMembers, &league.CreatedAt, &league.UpdatedAt, &league.MemberCount,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find mini league by id: %w", err)
	}
	return &league, nil
}

// FindByInviteCode returns a mini league by its invite code.
func (r *MiniLeagueRepo) FindByInviteCode(ctx context.Context, code string) (*domain.MiniLeague, error) {
	query := `
		SELECT ml.id, ml.name, ml.creator_id, ml.invite_code, ml.event_id, ml.max_members,
		       ml.created_at, ml.updated_at,
		       (SELECT COUNT(*) FROM mini_league_members WHERE league_id = ml.id) AS member_count
		FROM mini_leagues ml
		WHERE ml.invite_code = $1`

	var league domain.MiniLeague
	err := r.db.Pool.QueryRow(ctx, query, code).Scan(
		&league.ID, &league.Name, &league.CreatorID, &league.InviteCode, &league.EventID,
		&league.MaxMembers, &league.CreatedAt, &league.UpdatedAt, &league.MemberCount,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find mini league by invite code: %w", err)
	}
	return &league, nil
}

// FindByUser returns all mini leagues a user is a member of.
func (r *MiniLeagueRepo) FindByUser(ctx context.Context, userID uuid.UUID) ([]domain.MiniLeague, error) {
	query := `
		SELECT ml.id, ml.name, ml.creator_id, ml.invite_code, ml.event_id, ml.max_members,
		       ml.created_at, ml.updated_at,
		       (SELECT COUNT(*) FROM mini_league_members WHERE league_id = ml.id) AS member_count
		FROM mini_leagues ml
		JOIN mini_league_members mlm ON mlm.league_id = ml.id
		WHERE mlm.user_id = $1
		ORDER BY mlm.joined_at DESC`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("find leagues by user: %w", err)
	}
	defer rows.Close()

	var leagues []domain.MiniLeague
	for rows.Next() {
		var league domain.MiniLeague
		if err := rows.Scan(
			&league.ID, &league.Name, &league.CreatorID, &league.InviteCode, &league.EventID,
			&league.MaxMembers, &league.CreatedAt, &league.UpdatedAt, &league.MemberCount,
		); err != nil {
			return nil, fmt.Errorf("scan mini league: %w", err)
		}
		leagues = append(leagues, league)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mini leagues: %w", err)
	}
	return leagues, nil
}

// AddMember adds a user to a mini league.
func (r *MiniLeagueRepo) AddMember(ctx context.Context, leagueID, userID uuid.UUID) error {
	query := `
		INSERT INTO mini_league_members (id, league_id, user_id, joined_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (league_id, user_id) DO NOTHING`

	_, err := r.db.Pool.Exec(ctx, query, uuid.New(), leagueID, userID)
	if err != nil {
		return fmt.Errorf("add member to mini league: %w", err)
	}
	return nil
}

// RemoveMember removes a user from a mini league.
func (r *MiniLeagueRepo) RemoveMember(ctx context.Context, leagueID, userID uuid.UUID) error {
	query := `DELETE FROM mini_league_members WHERE league_id = $1 AND user_id = $2`

	ct, err := r.db.Pool.Exec(ctx, query, leagueID, userID)
	if err != nil {
		return fmt.Errorf("remove member from mini league: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("member not found in league")
	}
	return nil
}

// GetMembers returns all members of a mini league.
func (r *MiniLeagueRepo) GetMembers(ctx context.Context, leagueID uuid.UUID) ([]domain.MiniLeagueMember, error) {
	query := `
		SELECT id, league_id, user_id, joined_at
		FROM mini_league_members
		WHERE league_id = $1
		ORDER BY joined_at ASC`

	rows, err := r.db.Pool.Query(ctx, query, leagueID)
	if err != nil {
		return nil, fmt.Errorf("get mini league members: %w", err)
	}
	defer rows.Close()

	var members []domain.MiniLeagueMember
	for rows.Next() {
		var m domain.MiniLeagueMember
		if err := rows.Scan(&m.ID, &m.LeagueID, &m.UserID, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan mini league member: %w", err)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate mini league members: %w", err)
	}
	return members, nil
}

// CountMembers returns the number of members in a mini league.
func (r *MiniLeagueRepo) CountMembers(ctx context.Context, leagueID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM mini_league_members WHERE league_id = $1`

	var count int
	err := r.db.Pool.QueryRow(ctx, query, leagueID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count mini league members: %w", err)
	}
	return count, nil
}
