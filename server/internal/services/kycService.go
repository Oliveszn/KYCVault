package services

import (
	"context"
	"errors"
	"kycvault/internal/dtos"
	"kycvault/internal/models"
	"kycvault/internal/repository"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrSessionNotFound         = errors.New("Kyc session not found")
	ErrSessionAlreadyActive    = errors.New("You already have an active kyc session in progress")
	ErrSessionNotOwned         = errors.New("You do not have access to this session")
	ErrInvalidStatusTransition = errors.New("This action is not allowed at the current stage")
	ErrSessionAlreadyTerminal  = errors.New("This session has already been completed")
)

var allowedTransitions = map[models.KYCStatus][]models.KYCStatus{
	models.KYCStatusInitiated:  {models.KYCStatusDocUpload},
	models.KYCStatusDocUpload:  {models.KYCStatusFaceVerify},
	models.KYCStatusFaceVerify: {models.KYCStatusInReview},
	models.KYCStatusInReview:   {models.KYCStatusApproved, models.KYCStatusRejected},
	// Terminal states no outbound transitions.
	models.KYCStatusApproved: {},
	models.KYCStatusRejected: {},
}

type KYCService interface {
	// User
	InitiateSession(ctx context.Context, userID uuid.UUID, dto dtos.InitiateSessionRequest) (*models.KYCSession, error)
	GetSession(ctx context.Context, sessionID uuid.UUID) (*models.KYCSession, error)
	GetSessionForUser(ctx context.Context, sessionID, userID uuid.UUID) (*models.KYCSession, error)
	GetActiveSessionForUser(ctx context.Context, userID uuid.UUID) (*models.KYCSession, error)
	GetSessionHistoryForUser(ctx context.Context, userID uuid.UUID) ([]models.KYCSession, error)

	// Internal called by document/face/admin services, not directly by the handler
	AdvanceStatus(ctx context.Context, sessionID uuid.UUID, to models.KYCStatus, meta StatusMeta) error

	// Admin
	ApproveSession(ctx context.Context, sessionID, reviewerID uuid.UUID, note string) error
	RejectSession(ctx context.Context, sessionID, reviewerID uuid.UUID, note, reason string) error
	GetSessionQueue(ctx context.Context, limit, offset int) ([]models.KYCSession, int64, error)
	GetStatusCounts(ctx context.Context) (map[models.KYCStatus]int64, error)
}

// StatusMeta carries optional extra fields to persist alongside a status transition.
// Callers only populate the fields relevant to their transition.
type StatusMeta struct {
	VendorName      string
	VendorSessionID string
	RejectionReason string
	ReviewerID      *uuid.UUID
	ReviewNote      string
}

type kycService struct {
	repo   repository.KYCRepository
	audit  AuditService
	logger *zap.Logger
}

func NewKYCService(
	repo repository.KYCRepository,
	audit AuditService,
	logger *zap.Logger,
) KYCService {
	return &kycService{
		repo:   repo,
		audit:  audit,
		logger: logger,
	}
}

// InitiateSession creates a new KYC session for a user.
// It enforces that a user cannot have more than one active session at a time.
func (s *kycService) InitiateSession(ctx context.Context, userID uuid.UUID, dto dtos.InitiateSessionRequest) (*models.KYCSession, error) {
	expiresAt := time.Now().UTC().Add(24 * time.Hour) //24 hours rom creation

	//get the session history
	sessions, err := s.repo.GetSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, ErrInternal
	}

	attempt := 1

	if len(sessions) > 0 {
		lastSession := sessions[0]

		// Prevent multiple active sessions
		if !isTerminalStatus(lastSession.Status) {
			return nil, ErrSessionAlreadyActive
		}

		// Increment attempt
		attempt = lastSession.AttemptNumber + 1
	}

	session := &models.KYCSession{
		UserID:        userID,
		Status:        models.KYCStatusInitiated,
		Country:       dto.Country,
		IDType:        models.IDType(dto.IDType),
		AttemptNumber: attempt,
		ExpiresAt:     &expiresAt,
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		if errors.Is(err, repository.ErrSessionAlreadyActive) {
			return nil, ErrSessionAlreadyActive
		}
		s.logger.Error("failed to create kyc session",
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	s.audit.Log(ctx, models.AuditEvent{
		ActorID:   &userID,
		ActorRole: "user",
		SessionID: &session.ID,
		UserID:    &userID,
		EventType: models.AuditEventSessionCreated,
		// Metadata:  mustJSON(map[string]any{"country": dto.Country, "id_type": dto.IDType}),
		Metadata: map[string]any{"country": dto.Country, "id_type": dto.IDType},
	})

	s.logger.Info("kyc session initiated",
		zap.String("session_id", session.ID.String()),
		zap.String("user_id", userID.String()),
	)
	return session, nil
}

// GetSession fetches any session by ID. No ownership check — for internal use.
func (s *kycService) GetSession(ctx context.Context, sessionID uuid.UUID) (*models.KYCSession, error) {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, ErrInternal
	}
	return session, nil
}

// GetSessionForUser fetches a session and verifies it belongs to the requesting user.
// Use this in every user facing handler never GetSession directly.
func (s *kycService) GetSessionForUser(ctx context.Context, sessionID, userID uuid.UUID) (*models.KYCSession, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != userID {
		// Return NotFound rather than Forbidden don't confirm the session exists.
		return nil, ErrSessionNotFound
	}
	return session, nil
}

// GetActiveSessionForUser returns the user's current in-progress session, if any.
func (s *kycService) GetActiveSessionForUser(ctx context.Context, userID uuid.UUID) (*models.KYCSession, error) {
	session, err := s.repo.GetActiveSessionByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, ErrInternal
	}
	return session, nil
}

// GetSessionHistoryForUser returns all sessions for a user, terminal and active.
func (s *kycService) GetSessionHistoryForUser(ctx context.Context, userID uuid.UUID) ([]models.KYCSession, error) {
	s.logger.Info("fetching sessions",
		zap.String("user_id", userID.String()),
	)
	sessions, err := s.repo.GetSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, ErrInternal
	}
	return sessions, nil
}

// AdvanceStatus is the gatekeeper for all status transitions.
// Every other service that needs to move a session forward calls this, never
// writes to the status column directly. It validates legality, persists,
// and writes an audit event.
func (s *kycService) AdvanceStatus(ctx context.Context, sessionID uuid.UUID, to models.KYCStatus, meta StatusMeta) error {
	session, err := s.repo.GetSessionByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return ErrSessionNotFound
		}
		return ErrInternal
	}

	if err := s.validateTransition(session.Status, to); err != nil {
		return err
	}

	// fields := buildUpdateFields(to, meta)

	// if err := s.repo.UpdateSessionStatus(ctx, sessionID, to, fields); err != nil {
	// 	s.logger.Error("failed to advance session status",
	// 		zap.String("session_id", sessionID.String()),
	// 		zap.String("from", string(session.Status)),
	// 		zap.String("to", string(to)),
	// 		zap.Error(err),
	// 	)
	// 	return ErrInternal
	// }
	updated, err := s.repo.AdvanceStatusIfCurrent(
		ctx,
		sessionID,
		session.Status,
		to,
	)
	if err != nil {
		s.logger.Error("failed to advance session status",
			zap.String("session_id", sessionID.String()),
			zap.Error(err),
		)
		return ErrInternal
	}

	if !updated {
		return nil
	}

	s.audit.Log(ctx, models.AuditEvent{
		SessionID: &sessionID,
		UserID:    &session.UserID,
		ActorRole: "system",
		EventType: models.AuditEventStatusChanged,
		// Metadata: mustJSON(map[string]any{
		// 	"from":              string(session.Status),
		// 	"to":                string(to),
		// 	"vendor_name":       meta.VendorName,
		// 	"vendor_session_id": meta.VendorSessionID,
		// }),
		Metadata: map[string]any{
			"from":              session.Status,
			"to":                to,
			"vendor_name":       meta.VendorName,
			"vendor_session_id": meta.VendorSessionID,
		},
	})

	s.logger.Info("kyc session status advanced",
		zap.String("session_id", sessionID.String()),
		zap.String("from", string(session.Status)),
		zap.String("to", string(to)),
	)
	return nil
}

// ApproveSession is called by an admin. It wraps AdvanceStatus with admin-specific
// metadata and writes a manual approval audit event.
func (s *kycService) ApproveSession(ctx context.Context, sessionID, reviewerID uuid.UUID, note string) error {
	now := time.Now().UTC()

	if err := s.AdvanceStatus(ctx, sessionID, models.KYCStatusApproved, StatusMeta{
		ReviewerID: &reviewerID,
		ReviewNote: note,
	}); err != nil {
		return err
	}

	// Overwrite the audit event with the admin-specific type so it's
	// distinguishable from system transitions in the audit log.
	s.audit.Log(ctx, models.AuditEvent{
		ActorID:   &reviewerID,
		ActorRole: "admin",
		SessionID: &sessionID,
		EventType: models.AuditEventManualApproval,
		// Metadata:  mustJSON(map[string]any{"review_note": note, "reviewed_at": now}),
		Metadata: map[string]any{"review_note": note, "reviewed_at": now},
	})

	return nil
}

// RejectSession is called by an admin. Reason is user note is internal.
func (s *kycService) RejectSession(ctx context.Context, sessionID, reviewerID uuid.UUID, note, reason string) error {
	now := time.Now().UTC()

	if err := s.AdvanceStatus(ctx, sessionID, models.KYCStatusRejected, StatusMeta{
		ReviewerID:      &reviewerID,
		ReviewNote:      note,
		RejectionReason: reason,
	}); err != nil {
		return err
	}

	s.audit.Log(ctx, models.AuditEvent{
		ActorID:   &reviewerID,
		ActorRole: "admin",
		SessionID: &sessionID,
		EventType: models.AuditEventManualRejection,
		// Metadata:  mustJSON(map[string]any{"reason": reason, "note": note, "reviewed_at": now}),
		Metadata: map[string]any{"reason": reason, "note": note, "reviewed_at": now},
	})

	return nil
}

// / GetSessionQueue returns sessions awaiting manual review, oldest first.
func (s *kycService) GetSessionQueue(ctx context.Context, limit, offset int) ([]models.KYCSession, int64, error) {
	sessions, total, err := s.repo.GetSessionsByStatus(ctx, models.KYCStatusInReview, limit, offset)
	if err != nil {
		return nil, 0, ErrInternal
	}
	return sessions, total, nil
}

// GetStatusCounts returns a count per status for the admin dashboard tiles.
func (s *kycService) GetStatusCounts(ctx context.Context) (map[models.KYCStatus]int64, error) {
	statuses := []models.KYCStatus{
		models.KYCStatusInitiated,
		models.KYCStatusDocUpload,
		models.KYCStatusFaceVerify,
		models.KYCStatusInReview,
		models.KYCStatusApproved,
		models.KYCStatusRejected,
	}

	counts := make(map[models.KYCStatus]int64, len(statuses))
	for _, status := range statuses {
		count, err := s.repo.CountSessionsByStatus(ctx, status)
		if err != nil {
			return nil, ErrInternal
		}
		counts[status] = count
	}
	return counts, nil
}

// HELPERS
func (s *kycService) validateTransition(from, to models.KYCStatus) error {
	allowed, ok := allowedTransitions[from]
	if !ok {
		return ErrInvalidStatusTransition
	}
	for _, a := range allowed {
		if a == to {
			return nil
		}
	}
	if from == models.KYCStatusApproved || from == models.KYCStatusRejected {
		return ErrSessionAlreadyTerminal
	}
	return ErrInvalidStatusTransition
}

// buildUpdateFields constructs the extra columns to update alongside the status.
func buildUpdateFields(to models.KYCStatus, meta StatusMeta) map[string]any {
	fields := make(map[string]any)

	if meta.VendorName != "" {
		fields["vendor_name"] = meta.VendorName
	}
	if meta.VendorSessionID != "" {
		fields["vendor_session_id"] = meta.VendorSessionID
	}
	if meta.RejectionReason != "" {
		fields["rejection_reason"] = meta.RejectionReason
	}
	if meta.ReviewerID != nil {
		fields["reviewer_id"] = meta.ReviewerID
		fields["review_note"] = meta.ReviewNote
		now := time.Now().UTC()
		fields["reviewed_at"] = now
	}

	return fields
}

func isTerminalStatus(status models.KYCStatus) bool {
	return status == models.KYCStatusApproved ||
		status == models.KYCStatusRejected
}
