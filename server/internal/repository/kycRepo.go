package repository

import (
	"context"
	"errors"
	"fmt"
	"kycvault/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrSessionNotFound         = errors.New("kyc session not found")
	ErrSessionAlreadyActive    = errors.New("user already has an active kyc session")
	ErrSessionNotOwned         = errors.New("session does not belong to this user")
	ErrInvalidStatusTransition = errors.New("invalid session status transition")
)

type KYCRepository interface {
	CreateSession(ctx context.Context, session *models.KYCSession) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*models.KYCSession, error)
	GetActiveSessionByUserID(ctx context.Context, userID uuid.UUID) (*models.KYCSession, error)
	GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.KYCSession, error)
	UpdateSessionStatus(ctx context.Context, id uuid.UUID, status models.KYCStatus, fields map[string]any) error
	AdvanceStatusIfCurrent(ctx context.Context, id uuid.UUID, from models.KYCStatus, to models.KYCStatus) (bool, error)

	// Admin queries
	GetSessionsByStatus(ctx context.Context, status models.KYCStatus, limit, offset int) ([]models.KYCSession, int64, error)
	GetSessionByVendorSessionID(ctx context.Context, vendorSessionID string) (*models.KYCSession, error)
	CountSessionsByStatus(ctx context.Context, status models.KYCStatus) (int64, error)

	// Expiry cleanup todo: call it with a backgorund job
	ExpireStaleSession(ctx context.Context, id uuid.UUID) error
}

type kycRepository struct {
	db *gorm.DB
}

func NewKYCRepository(db *gorm.DB) KYCRepository {
	return &kycRepository{db: db}
}

// Createsession creates a new session. The db partial unique index
// will reject duplicates if the user already has an active session
func (r *kycRepository) CreateSession(ctx context.Context, session *models.KYCSession) error {
	result := r.db.WithContext(ctx).Create(session)
	if result.Error != nil {
		if IsDuplicateKeyError(result.Error) {
			return ErrSessionAlreadyActive
		}
		return fmt.Errorf("repository: create kyc session: %w", result.Error)
	}
	return nil
}

// GetSessionByID fetches a session with its document prloaded
func (r *kycRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*models.KYCSession, error) {
	var session models.KYCSession
	result := r.db.WithContext(ctx).Preload("Documents").First(&session, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("Repository: get session by id: %w", result.Error)
	}
	return &session, nil

}

// GetActiveSessionByUserID returns the single nonterminal session for a user
func (r *kycRepository) GetActiveSessionByUserID(ctx context.Context, userID uuid.UUID) (*models.KYCSession, error) {
	var session models.KYCSession
	result := r.db.WithContext(ctx).
		Preload("Documents").
		Where("user_id = ? AND status NOT IN ?", userID, []models.KYCStatus{
			models.KYCStatusApproved,
			models.KYCStatusRejected,
		}).
		First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("repository: get active session: %w", result.Error)
	}
	return &session, nil
}

// GetSessionByUserId returns all sessions for a user, newest first, for user's history view
func (r *kycRepository) GetSessionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.KYCSession, error) {
	var sessions []models.KYCSession
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&sessions)
	if result.Error != nil {
		return nil, fmt.Errorf("Repository: get session by user id: %w", result.Error)
	}
	return sessions, nil
}

// UpdateSessionStatus changes a session to a mew status and updates a set of additonal fields
// like reviewer_id, review_note, reject_reason
func (r *kycRepository) UpdateSessionStatus(ctx context.Context, id uuid.UUID, status models.KYCStatus, fields map[string]any) error {
	updates := map[string]any{
		"status":     status,
		"updated_at": time.Now().UTC(),
	}
	for k, v := range fields {
		updates[k] = v
	}

	result := r.db.WithContext(ctx).
		Model(&models.KYCSession{}).
		Where("id = ?", id).
		Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("repository: update session status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrSessionNotFound
	}
	return nil
}

// Advancestatus if current is a way to to update the current status of the session
// I previously handled this in the doc-service (in-memory) which led to race conditions
func (r *kycRepository) AdvanceStatusIfCurrent(ctx context.Context, id uuid.UUID, from models.KYCStatus, to models.KYCStatus) (bool, error) {

	result := r.db.WithContext(ctx).
		Model(&models.KYCSession{}).
		Where("id = ? AND status = ?", id, from).
		Updates(map[string]any{
			"status":     to,
			"updated_at": time.Now().UTC(),
		})

	if result.Error != nil {
		return false, result.Error
	}

	return result.RowsAffected > 0, nil
}

// GetSessionsByStatus returns a paginated list of sessions in a given status. like IN_REVIEW sessions
func (r *kycRepository) GetSessionsByStatus(ctx context.Context, status models.KYCStatus, limit, offset int) ([]models.KYCSession, int64, error) {
	var sessions []models.KYCSession
	var total int64

	query := r.db.WithContext(ctx).Model(&models.KYCSession{}).Where("status = ?", status)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("repository: count sessions by status: %w", err)
	}

	result := query.
		Preload("Documents").
		Order("created_at ASC"). // oldest first
		Limit(limit).
		Offset(offset).
		Find(&sessions)
	if result.Error != nil {
		return nil, 0, fmt.Errorf("repository: get sessions by status: %w", result.Error)
	}
	return sessions, total, nil
}

// GetSessionByVendorSessionID looks up a session by the vendor's reference ID.
// Called when the vendor fires an async webhook callback to your system.
func (r *kycRepository) GetSessionByVendorSessionID(ctx context.Context, vendorSessionID string) (*models.KYCSession, error) {
	var session models.KYCSession
	result := r.db.WithContext(ctx).
		Where("vendor_session_id = ?", vendorSessionID).
		First(&session)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("repository: get session by vendor id: %w", result.Error)
	}
	return &session, nil
}

// CountSessionsByStatus returns the total count for a status.
func (r *kycRepository) CountSessionsByStatus(ctx context.Context, status models.KYCStatus) (int64, error) {
	var count int64
	result := r.db.WithContext(ctx).
		Model(&models.KYCSession{}).
		Where("status = ?", status).
		Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("repository: count sessions: %w", result.Error)
	}
	return count, nil
}

// ExpireStaleSession transitions a session to rejected with a system-generated
// rejection reason. Only moves sessions that are still in a non-terminal state.
func (r *kycRepository) ExpireStaleSession(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.KYCSession{}).
		Where("id = ? AND status NOT IN ?", id, []models.KYCStatus{
			models.KYCStatusApproved,
			models.KYCStatusRejected,
		}).
		Updates(map[string]any{
			"status":           models.KYCStatusRejected,
			"rejection_reason": "session expired before verification was completed",
			"updated_at":       time.Now().UTC(),
		})
	if result.Error != nil {
		return fmt.Errorf("repository: expire session: %w", result.Error)
	}
	return nil
}
