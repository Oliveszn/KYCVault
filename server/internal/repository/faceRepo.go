package repository

import (
	"context"
	"errors"
	"fmt"
	"kycvault/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrFaceVerificationNotFound = errors.New("face verification not found")
)

type FaceRepository interface {
	//UpsertVerification create the face_verification row on first attempt and updates it on every retry, attempt history is in audit_events
	UpsertVerification(ctx context.Context, fv *models.FaceVerification) error

	// GetBySessionID fetches the single face verification record for a session.
	GetBySessionID(ctx context.Context, sessionID uuid.UUID) (*models.FaceVerification, error)

	//GetById fetches by primary key. used when correlating a face++ webhook callbak, we use FaceVerification.ID as the Face++ job ID
	GetByID(ctx context.Context, id uuid.UUID) (*models.FaceVerification, error)

	// UpdateResult writes the outcome fields after the face++ callback arrives.
	UpdateResult(ctx context.Context, id uuid.UUID, fields map[string]any) error

	// IncrementAttemptCount bumps the counter when a user retries face capture.
	IncrementAttemptCount(ctx context.Context, id uuid.UUID) error
}

type faceRepository struct {
	db *gorm.DB
}

func NewFaceRepository(db *gorm.DB) FaceRepository {
	return &faceRepository{db: db}
}

// Upsert creates or updates a row
func (r *faceRepository) UpsertVerification(ctx context.Context, fv *models.FaceVerification) error {
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "session_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"status",
				"selfie_storage_key",
				"selfie_storage_bucket",
				"selfie_checksum",
				"vendor_name",
				"vendor_request_id",
				"attempt_count",
				"failure_reason",
				"liveness_score",
				"liveness_passed",
				"match_score",
				"match_threshold",
				"match_passed",
				"vendor_raw_result",
				"updated_at",
			}),
		}).
		Create(fv)

	if result.Error != nil {
		return fmt.Errorf("repository: upsert face verification: %w", result.Error)
	}
	return nil
}

// GetBySessionID returns the face verification for a session.
func (r *faceRepository) GetBySessionID(ctx context.Context, sessionID uuid.UUID) (*models.FaceVerification, error) {
	var fv models.FaceVerification
	result := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		First(&fv)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrFaceVerificationNotFound
		}
		return nil, fmt.Errorf("repository: get face verification by session: %w", result.Error)
	}
	return &fv, nil
}

// GetByID fetches by primary key, this is how we look up which session a callback belongs to.
func (r *faceRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.FaceVerification, error) {
	var fv models.FaceVerification
	result := r.db.WithContext(ctx).
		First(&fv, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrFaceVerificationNotFound
		}
		return nil, fmt.Errorf("repository: get face verification by id: %w", result.Error)
	}
	return &fv, nil
}

// UpdateResult writes the outcome fields that arrive via the facepp.
func (r *faceRepository) UpdateResult(ctx context.Context, id uuid.UUID, fields map[string]any) error {
	fields["updated_at"] = time.Now().UTC()

	result := r.db.WithContext(ctx).
		Model(&models.FaceVerification{}).
		Where("id = ?", id).
		Updates(fields)
	if result.Error != nil {
		return fmt.Errorf("repository: update face verification result: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrFaceVerificationNotFound
	}
	return nil
}

// IncrementAttemptCount uses a single UPDATE
// makes it safe under concurrent requests no read-then-write race condition, unlike handling it in-memory
func (r *faceRepository) IncrementAttemptCount(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.FaceVerification{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"attempt_count": gorm.Expr("attempt_count + 1"),
			"updated_at":    time.Now().UTC(),
		})
	if result.Error != nil {
		return fmt.Errorf("repository: increment attempt count: %w", result.Error)
	}
	return nil
}
