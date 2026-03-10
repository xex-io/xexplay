package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/xex-exchange/xexplay-api/internal/domain"
	"github.com/xex-exchange/xexplay-api/internal/repository/postgres"
)

type AuditService struct {
	auditRepo *postgres.AuditRepo
}

func NewAuditService(auditRepo *postgres.AuditRepo) *AuditService {
	return &AuditService{auditRepo: auditRepo}
}

// LogAction records an admin action in the audit log.
func (s *AuditService) LogAction(ctx context.Context, adminUserID uuid.UUID, action, entityType, entityID string, details interface{}, ipAddress string) {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		log.Error().Err(err).Str("action", action).Msg("failed to marshal audit details")
		detailsJSON = []byte(`{}`)
	}

	entry := &domain.AuditLog{
		ID:          uuid.New(),
		AdminUserID: adminUserID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		Details:     detailsJSON,
		IPAddress:   ipAddress,
	}

	if err := s.auditRepo.Create(ctx, entry); err != nil {
		log.Error().Err(err).Str("action", action).Msg("failed to create audit log")
	}
}

// GetRecentLogs returns recent audit log entries with pagination.
func (s *AuditService) GetRecentLogs(ctx context.Context, limit, offset int) ([]domain.AuditLog, error) {
	logs, err := s.auditRepo.FindRecent(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get recent audit logs: %w", err)
	}
	return logs, nil
}

// GetLogsByAdmin returns audit logs for a specific admin user.
func (s *AuditService) GetLogsByAdmin(ctx context.Context, adminUserID uuid.UUID, limit, offset int) ([]domain.AuditLog, error) {
	logs, err := s.auditRepo.FindByAdmin(ctx, adminUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get audit logs by admin: %w", err)
	}
	return logs, nil
}

// GetLogsByEntity returns audit logs for a specific entity.
func (s *AuditService) GetLogsByEntity(ctx context.Context, entityType, entityID string, limit, offset int) ([]domain.AuditLog, error) {
	logs, err := s.auditRepo.FindByEntity(ctx, entityType, entityID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get audit logs by entity: %w", err)
	}
	return logs, nil
}
