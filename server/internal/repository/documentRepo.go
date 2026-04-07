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
	ErrDocumentNotFound   = errors.New("document not found")
	ErrDocumentSideExists = errors.New("an accepted document for this side already exists")
	ErrDocumentNotOwned   = errors.New("document does not belong to this session")
)

type DocumentRepository interface {
	CreateDocument(ctx context.Context, doc *models.KYCDocument) error
	GetDocumentByID(ctx context.Context, id uuid.UUID) (*models.KYCDocument, error)
	GetDocumentsBySession(ctx context.Context, sessionID uuid.UUID) ([]models.KYCDocument, error)
	UpdateDocumentStatus(ctx context.Context, id uuid.UUID, status models.DocumentStatus, fields map[string]any) error

	//Both side accepted returns true when the session has front and back accepted
	//KYC service calls this before moving session to face_verify
	BothSidesAccepted(ctx context.Context, sessionID uuid.UUID) (bool, error)

	//GetAcceptedDocument returns the accepted documnet for a given side
	//The face service uses this to retriev the reference image storage key, when running face match against document photo
	GetAcceptedDocument(ctx context.Context, sessionID uuid.UUID, side models.DocumentSide) (*models.KYCDocument, error)
}

type documentRepository struct {
	db *gorm.DB
}

func NewDocumentRepository(db *gorm.DB) DocumentRepository {
	return &documentRepository{db: db}
}

// CreateDocument inserts a new document record, the partial unique index
// idx_kyc_documents_one_accepted_per_side prevents a second accepted record for the smae side on the session, violation is translated here
func (r *documentRepository) CreateDocument(ctx context.Context, doc *models.KYCDocument) error {
	result := r.db.WithContext(ctx).Create(doc)
	if result.Error != nil {
		if IsDuplicateKeyError(result.Error) {
			return ErrDocumentSideExists
		}
		return fmt.Errorf("repository: create document: %w", result.Error)
	}
	return nil
}

// GetDocumentByID fetches a single document by its primary key.
func (r *documentRepository) GetDocumentByID(ctx context.Context, id uuid.UUID) (*models.KYCDocument, error) {
	var doc models.KYCDocument
	result := r.db.WithContext(ctx).
		First(&doc, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, fmt.Errorf("repository: get document by id: %w", result.Error)
	}
	return &doc, nil
}

// GetDocumentsBySession returns all documents for a session, ordered by upload time.
func (r *documentRepository) GetDocumentsBySession(ctx context.Context, sessionID uuid.UUID) ([]models.KYCDocument, error) {
	var docs []models.KYCDocument
	result := r.db.WithContext(ctx).
		Where("session_id = ?", sessionID).
		Order("uploaded_at ASC").
		Find(&docs)
	if result.Error != nil {
		return nil, fmt.Errorf("repository: get documents by session: %w", result.Error)
	}
	return docs, nil
}

// UpdateDocumentStatus updates the status column and any extra fields passed in the fields map like extracted_data, rejection_reason
func (r *documentRepository) UpdateDocumentStatus(ctx context.Context, id uuid.UUID, status models.DocumentStatus, fields map[string]any) error {
	updates := map[string]any{
		"status":     status,
		"updated_at": time.Now().UTC(),
	}

	for k, v := range fields {
		updates[k] = v
	}

	result := r.db.WithContext(ctx).Model(&models.KYCDocument{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return fmt.Errorf("repository: update document status: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrDocumentNotFound
	}
	return nil
}

// this retunrs true when the session has one accepted front and back, we use a count query so it never loads image data
func (r *documentRepository) BothSidesAccepted(ctx context.Context, sessionID uuid.UUID) (bool, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&models.KYCDocument{}).Where("session_id = ? AND status = ? AND side IN ?",
		sessionID,
		models.DocumentStatusAccepted,
		[]models.DocumentSide{models.DocumentSideFront, models.DocumentSideBack}).
		Distinct("side").
		Count(&count)
	if result.Error != nil {
		return false, fmt.Errorf("repository: both sides accepted: %w", result.Error)
	}
	return count == 2, nil
}

// this returns the single accepted document for the given side.
// Returns ErrDocumentNotFound if no accepted document exists yet for that side.
func (r *documentRepository) GetAcceptedDocument(ctx context.Context, sessionID uuid.UUID, side models.DocumentSide) (*models.KYCDocument, error) {
	var doc models.KYCDocument
	result := r.db.WithContext(ctx).
		Where("session_id = ? AND side = ? AND status = ?",
			sessionID, side, models.DocumentStatusAccepted,
		).
		First(&doc)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, fmt.Errorf("repository: get accepted document: %w", result.Error)
	}
	return &doc, nil
}
