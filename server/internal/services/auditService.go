package services

import (
	"context"
	"encoding/json"
	"kycvault/internal/models"
	"kycvault/internal/repository"

	"go.uber.org/zap"
)

// AuditService is the single write path for the append-only audit_events table.
// All other services call this — nothing writes to audit_events directly.
type AuditService interface {
	Log(ctx context.Context, event models.AuditEvent)
}

type auditService struct {
	repo   repository.AuditRepository
	logger *zap.Logger
}

func NewAuditService(repo repository.AuditRepository, logger *zap.Logger) AuditService {
	return &auditService{repo: repo, logger: logger}
}

// Log inserts an audit event. It is fire-and-forget — a logging failure must
// never bubble up and break the caller's happy path. Errors are logged internally.
func (s *auditService) Log(ctx context.Context, event models.AuditEvent) {
	if err := s.repo.CreateAuditEvent(ctx, &event); err != nil {
		s.logger.Error("failed to write audit event",
			zap.String("event_type", string(event.EventType)),
			zap.Error(err),
		)
	}
}

// These live here so every service file can import them without a cycle.

// mustJSON is the service-layer equivalent used when building audit Metadata.
func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
