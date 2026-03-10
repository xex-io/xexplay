package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/xex-exchange/xexplay-api/internal/domain"
)

type AchievementRepo struct {
	db *DB
}

func NewAchievementRepo(db *DB) *AchievementRepo {
	return &AchievementRepo{db: db}
}

// FindAll returns all defined achievements.
func (r *AchievementRepo) FindAll(ctx context.Context) ([]domain.Achievement, error) {
	query := `
		SELECT id, key, name, description, icon, category, condition_type, condition_value, created_at
		FROM achievements
		ORDER BY category, key`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("find all achievements: %w", err)
	}
	defer rows.Close()

	var achievements []domain.Achievement
	for rows.Next() {
		var a domain.Achievement
		if err := rows.Scan(
			&a.ID, &a.Key, &a.Name, &a.Description, &a.Icon, &a.Category,
			&a.ConditionType, &a.ConditionValue, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan achievement: %w", err)
		}
		achievements = append(achievements, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate achievements: %w", err)
	}
	return achievements, nil
}

// FindByUser returns all achievements earned by a user, including achievement details.
func (r *AchievementRepo) FindByUser(ctx context.Context, userID uuid.UUID) ([]domain.UserAchievement, error) {
	query := `
		SELECT ua.id, ua.user_id, ua.achievement_id, ua.earned_at,
		       a.id, a.key, a.name, a.description, a.icon, a.category,
		       a.condition_type, a.condition_value, a.created_at
		FROM user_achievements ua
		JOIN achievements a ON a.id = ua.achievement_id
		WHERE ua.user_id = $1
		ORDER BY ua.earned_at DESC`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("find user achievements: %w", err)
	}
	defer rows.Close()

	var results []domain.UserAchievement
	for rows.Next() {
		var ua domain.UserAchievement
		var a domain.Achievement
		if err := rows.Scan(
			&ua.ID, &ua.UserID, &ua.AchievementID, &ua.EarnedAt,
			&a.ID, &a.Key, &a.Name, &a.Description, &a.Icon, &a.Category,
			&a.ConditionType, &a.ConditionValue, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan user achievement: %w", err)
		}
		ua.Achievement = &a
		results = append(results, ua)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user achievements: %w", err)
	}
	return results, nil
}

// Grant inserts a user_achievement record.
func (r *AchievementRepo) Grant(ctx context.Context, userID, achievementID uuid.UUID) error {
	query := `
		INSERT INTO user_achievements (id, user_id, achievement_id, earned_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id, achievement_id) DO NOTHING`

	_, err := r.db.Pool.Exec(ctx, query, uuid.New(), userID, achievementID)
	if err != nil {
		return fmt.Errorf("grant achievement: %w", err)
	}
	return nil
}

// HasAchievement checks if a user already has a specific achievement.
func (r *AchievementRepo) HasAchievement(ctx context.Context, userID, achievementID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM user_achievements
			WHERE user_id = $1 AND achievement_id = $2
		)`

	var exists bool
	err := r.db.Pool.QueryRow(ctx, query, userID, achievementID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check achievement: %w", err)
	}
	return exists, nil
}

// FindByConditionType returns achievements matching a given condition type.
func (r *AchievementRepo) FindByConditionType(ctx context.Context, conditionType string) ([]domain.Achievement, error) {
	query := `
		SELECT id, key, name, description, icon, category, condition_type, condition_value, created_at
		FROM achievements
		WHERE condition_type = $1`

	rows, err := r.db.Pool.Query(ctx, query, conditionType)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find achievements by condition type: %w", err)
	}
	defer rows.Close()

	var achievements []domain.Achievement
	for rows.Next() {
		var a domain.Achievement
		if err := rows.Scan(
			&a.ID, &a.Key, &a.Name, &a.Description, &a.Icon, &a.Category,
			&a.ConditionType, &a.ConditionValue, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan achievement: %w", err)
		}
		achievements = append(achievements, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate achievements: %w", err)
	}
	return achievements, nil
}
