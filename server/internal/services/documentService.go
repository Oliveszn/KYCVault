package services

import (
	"bytes"
	"context"

	"errors"
	"fmt"
	"io"
	"kycvault/internal/infra/storage"
	"kycvault/internal/models"
	"kycvault/internal/repository"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrDocumentNotFound      = errors.New("Document not found")
	ErrDocumentNotOwned      = errors.New("Document does not belong to this session")
	ErrInvalidFileType       = errors.New("File type not allowed; accepted types: jpeg, png, pdf")
	ErrFileTooLarge          = errors.New("File exceeds the maximum allowed size of 10 MB")
	ErrInvalidDocumentSide   = errors.New("Side must be 'front' or 'back'")
	ErrSessionWrongStage     = errors.New("Session is not at the document upload stage")
	ErrDocumentSetIncomplete = errors.New("Both front and back documents must be accepted before proceeding")
)

const (
	maxFileSizeBytes = 10 << 20 // 10 MB
	presignedURLTTL  = 15 * time.Minute
)

var allowedMIMETypes = map[string]string{
	"image/jpeg":      "jpg",
	"image/png":       "png",
	"application/pdf": "pdf",
}

type DocumentService interface {
	// UploadDocument is called by the handler after parsing a multipart upload.
	// It validates, stores, and if both sides are now accepted advances the session to face_verify.
	UploadDocument(ctx context.Context, req UploadDocumentRequest) (*models.KYCDocument, error)

	// GetDocumentsForSession returns all documents for a session.
	GetDocumentsForSession(ctx context.Context, sessionID, userID uuid.UUID) ([]models.KYCDocument, error)

	// GetPresignedURL returns a short-lived URL for viewing a document image, for admins only
	GetPresignedURL(ctx context.Context, docID uuid.UUID) (string, error)
}

// uploaddocumentrequest carries all data the handler collects from the multipart request before handing off to the service
type UploadDocumentRequest struct {
	SessionID  uuid.UUID
	UserID     uuid.UUID
	Side       string
	FileHeader *multipart.FileHeader
	IPAddress  string
	UserAgent  string
}

type documentService struct {
	docRepo repository.DocumentRepository
	kycRepo repository.KYCRepository
	kycSvc  KYCService
	storage storage.Client
	bucket  string
	audit   AuditService
	logger  *zap.Logger
}

func NewDocumentService(
	docRepo repository.DocumentRepository,
	kycRepo repository.KYCRepository,
	kycSvc KYCService,
	storageClient storage.Client,
	bucket string,
	audit AuditService,
	logger *zap.Logger,
) DocumentService {
	return &documentService{
		docRepo: docRepo,
		kycRepo: kycRepo,
		kycSvc:  kycSvc,
		storage: storageClient,
		bucket:  bucket,
		audit:   audit,
		logger:  logger,
	}
}

func (s *documentService) UploadDocument(ctx context.Context, req UploadDocumentRequest) (*models.KYCDocument, error) {
	s.logger.Info("upload document started",
		zap.String("session_id", req.SessionID.String()),
		zap.String("user_id", req.UserID.String()),
		zap.String("side_raw", req.Side),
	)

	// Validate the document side
	side, err := parseDocumentSide(req.Side)
	if err != nil {
		s.logger.Warn("invalid document side",
			zap.String("side", req.Side),
			zap.Error(err),
		)
		return nil, err
	}

	// Verify session ownership
	session, err := s.kycSvc.GetSessionForUser(ctx, req.SessionID, req.UserID)
	if err != nil {
		s.logger.Warn("session validation failed",
			zap.String("session_id", req.SessionID.String()),
			zap.String("user_id", req.UserID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	// Validate session stage
	if session.Status != models.KYCStatusInitiated && session.Status != models.KYCStatusDocUpload {
		s.logger.Warn("invalid session stage for upload",
			zap.String("session_id", session.ID.String()),
			zap.String("status", string(session.Status)),
		)
		return nil, ErrSessionWrongStage
	}

	// Validate file
	mimeType, ext, err := validateFile(req.FileHeader)
	if err != nil {
		s.logger.Warn("invalid file upload",
			zap.String("filename", req.FileHeader.Filename),
			zap.Error(err),
		)
		return nil, err
	}

	// Read file
	file, err := req.FileHeader.Open()
	if err != nil {
		s.logger.Error("failed to open uploaded file", zap.Error(err))
		return nil, fmt.Errorf("service: open uploaded file: %w", err)
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(io.LimitReader(file, maxFileSizeBytes+1))
	if err != nil {
		s.logger.Error("failed to read uploaded file", zap.Error(err))
		return nil, fmt.Errorf("service: read uploaded file: %w", err)
	}

	if int64(len(fileBytes)) > maxFileSizeBytes {
		s.logger.Warn("file too large",
			zap.Int64("size", int64(len(fileBytes))),
		)
		return nil, ErrFileTooLarge
	}

	checksum := computeChecksum(fileBytes)

	// Build document
	docID := uuid.New()
	storageKey := storage.KeyForDocument(req.SessionID.String(), docID.String(), string(side), ext)

	doc := &models.KYCDocument{
		ID:               docID,
		SessionID:        req.SessionID,
		UserID:           req.UserID,
		Side:             side,
		Status:           models.DocumentStatusPending,
		StorageKey:       storageKey,
		StorageBucket:    s.bucket,
		OriginalFilename: req.FileHeader.Filename,
		MIMEType:         mimeType,
		FileSizeBytes:    int64(len(fileBytes)),
		Checksum:         checksum,
	}

	// Save record
	if err := s.docRepo.CreateDocument(ctx, doc); err != nil {
		if errors.Is(err, repository.ErrDocumentSideExists) {
			s.logger.Warn("document side already exists",
				zap.String("session_id", req.SessionID.String()),
				zap.String("side", string(side)),
			)
			return nil, repository.ErrDocumentSideExists
		}

		s.logger.Error("failed to create document record",
			zap.String("session_id", req.SessionID.String()),
			zap.String("side", string(side)),
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	// Upload to storage
	if err := s.storage.Put(
		ctx,
		s.bucket,
		storageKey,
		bytes.NewReader(fileBytes),
		int64(len(fileBytes)),
		mimeType,
	); err != nil {

		_ = s.docRepo.UpdateDocumentStatus(ctx, doc.ID, models.DocumentStatusRejected, map[string]any{
			"rejection_reason": "storage upload failed; please re-upload",
		})

		s.logger.Error("storage upload failed",
			zap.String("doc_id", doc.ID.String()),
			zap.String("key", storageKey),
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	// Accept document
	if err := s.docRepo.UpdateDocumentStatus(ctx, doc.ID, models.DocumentStatusAccepted, nil); err != nil {
		s.logger.Error("failed to mark document as accepted",
			zap.String("doc_id", doc.ID.String()),
			zap.Error(err),
		)
		return nil, ErrInternal
	}
	doc.Status = models.DocumentStatusAccepted

	// Audit log
	s.audit.Log(ctx, models.AuditEvent{
		ActorID:   &req.UserID,
		ActorRole: "user",
		SessionID: &req.SessionID,
		UserID:    &req.UserID,
		EventType: models.AuditEventDocumentUploaded,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
		Metadata: map[string]any{
			"doc_id":    doc.ID,
			"side":      string(side),
			"mime_type": mimeType,
			"size":      doc.FileSizeBytes,
		},
	})

	// Advance session
	if err := s.advanceSessionIfReady(ctx, session, req.UserID); err != nil {
		s.logger.Error("failed to advance session after upload",
			zap.String("session_id", req.SessionID.String()),
			zap.Error(err),
		)
	}

	s.logger.Info("document upload completed successfully",
		zap.String("doc_id", doc.ID.String()),
		zap.String("session_id", req.SessionID.String()),
		zap.String("user_id", req.UserID.String()),
		zap.String("side", string(side)),
	)

	return doc, nil
}

func (s *documentService) GetDocumentsForSession(ctx context.Context, sessionID, userID uuid.UUID) ([]models.KYCDocument, error) {
	s.logger.Info("fetching documents for session",
		zap.String("session_id", sessionID.String()),
		zap.String("user_id", userID.String()),
	)

	if _, err := s.kycSvc.GetSessionForUser(ctx, sessionID, userID); err != nil {
		s.logger.Warn("session ownership validation failed",
			zap.String("session_id", sessionID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	docs, err := s.docRepo.GetDocumentsBySession(ctx, sessionID)
	if err != nil {
		s.logger.Error("failed to fetch documents",
			zap.String("session_id", sessionID.String()),
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	s.logger.Info("documents fetched successfully",
		zap.String("session_id", sessionID.String()),
		zap.Int("count", len(docs)),
	)

	return docs, nil
}

func (s *documentService) GetPresignedURL(ctx context.Context, docID uuid.UUID) (string, error) {
	s.logger.Info("generating presigned url",
		zap.String("doc_id", docID.String()),
	)

	doc, err := s.docRepo.GetDocumentByID(ctx, docID)
	if err != nil {
		if errors.Is(err, repository.ErrDocumentNotFound) {
			s.logger.Warn("document not found",
				zap.String("doc_id", docID.String()),
			)
			return "", ErrDocumentNotFound
		}

		s.logger.Error("failed to fetch document",
			zap.String("doc_id", docID.String()),
			zap.Error(err),
		)
		return "", ErrInternal
	}

	url, err := s.storage.GetPresignedURL(ctx, doc.StorageBucket, doc.StorageKey, presignedURLTTL)
	if err != nil {
		s.logger.Error("failed to generate presigned url",
			zap.String("doc_id", docID.String()),
			zap.Error(err),
		)
		return "", ErrInternal
	}

	s.logger.Info("presigned url generated successfully",
		zap.String("doc_id", docID.String()),
	)

	return url, nil
}

// HELPERS

// advanceSessionIfReady checks whether both document sides are now accepted
// and, if so, advances the session from doc_upload to face_verify.
func (s *documentService) advanceSessionIfReady(ctx context.Context, session *models.KYCSession, userID uuid.UUID) error {
	s.logger.Info("advanceSessionIfReady called",
		zap.String("session_id", session.ID.String()),
		zap.String("current_status", string(session.Status)),
	)
	//ensure we are at least in doc_upload
	if session.Status == models.KYCStatusInitiated {
		if err := s.kycSvc.AdvanceStatus(
			ctx,
			session.ID,
			models.KYCStatusDocUpload,
			StatusMeta{},
		); err != nil {
			s.logger.Error("failed to advance to doc_upload",
				zap.String("session_id", session.ID.String()),
				zap.Error(err),
			)
			return err
		}

		session.Status = models.KYCStatusDocUpload
	}

	//check if both sides are done
	bothDone, err := s.docRepo.BothSidesAccepted(ctx, session.ID)
	if err != nil {
		s.logger.Error("both sides check failed",
			zap.String("session_id", session.ID.String()),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("both sides check",
		zap.Bool("both_done", bothDone),
	)

	if !bothDone {
		return nil
	}
	//advance to face_verify if still in doc_upload
	if err := s.kycSvc.AdvanceStatus(
		ctx,
		session.ID,
		models.KYCStatusFaceVerify,
		StatusMeta{},
	); err != nil {
		s.logger.Error("failed to advance to face_verify",
			zap.String("session_id", session.ID.String()),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func validateFile(fh *multipart.FileHeader) (mimeType, ext string, err error) {
	if fh.Size > maxFileSizeBytes {
		return "", "", ErrFileTooLarge
	}

	// Detect MIME type from the Content-Type header the browser sends.
	// For a production system you should also sniff the first 512 bytes
	// of the file itself with http.DetectContentType as a second check.
	ct := fh.Header.Get("Content-Type")
	// Strip any parameters (e.g. "image/jpeg; charset=utf-8")
	if idx := strings.Index(ct, ";"); idx != -1 {
		ct = strings.TrimSpace(ct[:idx])
	}

	ext, ok := allowedMIMETypes[ct]
	if !ok {
		// Fall back to extension check so "image/jpg" browsers don't break things.
		fileExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(fh.Filename), "."))
		for mime, e := range allowedMIMETypes {
			if e == fileExt {
				return mime, e, nil
			}
		}
		return "", "", ErrInvalidFileType
	}
	return ct, ext, nil
}

func parseDocumentSide(raw string) (models.DocumentSide, error) {
	switch models.DocumentSide(raw) {
	case models.DocumentSideFront, models.DocumentSideBack:
		return models.DocumentSide(raw), nil
	default:
		return "", ErrInvalidDocumentSide
	}
}
