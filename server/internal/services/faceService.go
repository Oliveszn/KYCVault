package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"kycvault/internal/infra/facepp"
	"kycvault/internal/infra/storage"
	"kycvault/internal/models"
	"kycvault/internal/repository"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrFaceVerificationNotFound = errors.New("face verification not found")
	ErrFaceSessionWrongStage    = errors.New("session is not at the face verification stage")
	ErrFaceMaxAttemptsReached   = errors.New("maximum face verification attempts reached")
	ErrFaceDocumentNotReady     = errors.New("document images are not available for face matching")
	ErrFaceUnsupportedIDType    = errors.New("id type is not supported for this country")
	ErrFaceVerificationPending  = errors.New("face verification is already in progress")
)

const (
	maxSelfieBytes  = 5 << 20 // set to 5mb selfies don't need to be as large as documents
	maxFaceAttempts = 3       // user gets 3 tries before the session is escalated to manual review
)

type FaceService interface {
	// StartVerification is called when the user submits their selfie.
	// It stores the image, submits the biometric job to facepp Identity,and returns immediately
	StartVerification(ctx context.Context, req StartVerificationRequest) (*models.FaceVerification, error)

	// GetVerification returns the face verification for a session.

	GetVerificationForUser(ctx context.Context, sessionID, userID uuid.UUID) (*models.FaceVerification, error)
}

// StartVerificationRequest carries the selfie data parsed by the handler.
type StartVerificationRequest struct {
	SessionID  uuid.UUID
	UserID     uuid.UUID
	FileHeader *multipart.FileHeader
	IPAddress  string
	UserAgent  string
}

type faceService struct {
	faceRepo repository.FaceRepository
	docRepo  repository.DocumentRepository
	kycSvc   KYCService
	facepp   facepp.Client
	storage  storage.Client
	bucket   string
	audit    AuditService
	logger   *zap.Logger
}

func NewFaceService(
	faceRepo repository.FaceRepository,
	docRepo repository.DocumentRepository,
	kycSvc KYCService,
	facepp facepp.Client,
	storageClient storage.Client,
	bucket string,
	audit AuditService,
	logger *zap.Logger,
) FaceService {
	return &faceService{
		faceRepo: faceRepo,
		docRepo:  docRepo,
		kycSvc:   kycSvc,
		facepp:   facepp,
		storage:  storageClient,
		bucket:   bucket,
		audit:    audit,
		logger:   logger,
	}
}

func (s *faceService) StartVerification(ctx context.Context, req StartVerificationRequest) (*models.FaceVerification, error) {
	// Load session and ownership check
	session, err := s.kycSvc.GetSessionForUser(ctx, req.SessionID, req.UserID)
	if err != nil {
		return nil, err
	}

	//
	if session.Status != models.KYCStatusFaceVerify {
		return nil, ErrFaceSessionWrongStage
	}

	// check if there's already a pending verification
	existing, err := s.faceRepo.GetBySessionID(ctx, req.SessionID)
	if err != nil && !errors.Is(err, repository.ErrFaceVerificationNotFound) {
		return nil, ErrInternal
	}
	if existing != nil && existing.Status == models.FaceVerificationStatusPending {
		return nil, ErrFaceVerificationPending
	}

	// enforcing max attempts
	if existing != nil && existing.AttemptCount >= maxFaceAttempts {
		return nil, ErrFaceMaxAttemptsReached
	}

	// validate and read the selfie
	selfieBytes, err := readSelfie(req.FileHeader)
	if err != nil {
		return nil, err
	}
	checksum := computeChecksum(selfieBytes)

	//load the front document for face match chech
	frontDoc, err := s.docRepo.GetAcceptedDocument(ctx, req.SessionID, models.DocumentSideFront)
	if err != nil {
		if errors.Is(err, repository.ErrDocumentNotFound) {
			return nil, ErrFaceDocumentNotReady
		}
		return nil, ErrInternal
	}

	//seting IDs up front to use as storage key facepp id
	verificationID := uuid.New()
	if existing != nil {
		// Reuse the same record ID on retry so there's always one row per session.
		verificationID = existing.ID
	}

	selfieKey := storage.KeyForSelfie(req.SessionID.String(), verificationID.String())

	//build and upsert the verification record before sending to vendor
	//if the upload to facpp fails we have a pending row so admin can inspect
	attemptCount := 1
	if existing != nil {
		attemptCount = existing.AttemptCount + 1
	}

	fv := &models.FaceVerification{
		ID:                  verificationID,
		SessionID:           req.SessionID,
		UserID:              req.UserID,
		Status:              models.FaceVerificationStatusPending,
		SelfieStorageKey:    selfieKey,
		SelfieStorageBucket: s.bucket,
		SelfieChecksum:      checksum,
		VendorName:          "facepp",
		VendorRequestID:     verificationID.String(), // we use our ID as the facepp ID
		AttemptCount:        attemptCount,
	}

	if err := s.faceRepo.UpsertVerification(ctx, fv); err != nil {
		s.logger.Error("failed to upsert face verification record",
			zap.String("session_id", req.SessionID.String()),
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	//store the selfie in object storage
	if err := s.storage.Put(
		ctx,
		s.bucket,
		selfieKey,
		bytes.NewReader(selfieBytes),
		int64(len(selfieBytes)),
		"image/jpeg",
	); err != nil {
		s.logger.Error("selfie storage upload failed",
			zap.String("verification_id", verificationID.String()),
			zap.Error(err),
		)
		// Mark failed so the user can retry — don't leave them stuck on pending.
		_ = s.faceRepo.UpdateResult(ctx, verificationID, map[string]any{
			"status":         models.FaceVerificationStatusFailed,
			"failure_reason": "selfie upload failed; please retry",
		})
		return nil, ErrInternal
	}

	// fetch the documnet and load

	docBytes, err := s.fetchStorageBytes(ctx, frontDoc.StorageBucket, frontDoc.StorageKey)
	if err != nil {
		s.logger.Error("failed to fetch front document for facepp submission",
			zap.String("doc_id", frontDoc.ID.String()),
			zap.Error(err),
		)
		return nil, ErrInternal
	}
	result, err := s.facepp.CompareFaces(ctx, selfieBytes, docBytes)
	if err != nil {
		s.logger.Error("face++ compare failed",
			zap.String("verification_id", verificationID.String()),
			zap.Error(err),
		)

		_ = s.faceRepo.UpdateResult(ctx, verificationID, map[string]any{
			"status":         models.FaceVerificationStatusFailed,
			"failure_reason": "face verification failed; please retry",
		})

		return nil, ErrInternal
	}

	// decide if pass or fail
	passed := result.Confidence >= 80.0

	status := models.FaceVerificationStatusFailed
	if passed {
		status = models.FaceVerificationStatusPassed
	}

	err = s.faceRepo.UpdateResult(ctx, verificationID, map[string]any{
		"status":            status,
		"match_score":       result.Confidence,
		"match_threshold":   80.0,
		"match_passed":      passed,
		"vendor_name":       "facepp",
		"vendor_raw_result": mustJSON(result),
	})

	if err != nil {
		return nil, ErrInternal
	}

	s.audit.Log(ctx, models.AuditEvent{
		ActorID:   &req.UserID,
		ActorRole: "user",
		SessionID: &req.SessionID,
		UserID:    &req.UserID,
		EventType: models.AuditEventFaceVerifyPassed,
		IPAddress: req.IPAddress,
		UserAgent: req.UserAgent,
		Metadata: mustJSON(map[string]any{
			"verification_id": verificationID,
			"attempt_count":   attemptCount,
			"vendor":          "facepp",
		}),
	})

	s.logger.Info("face verification submitted to facepp",
		zap.String("verification_id", verificationID.String()),
		zap.String("session_id", req.SessionID.String()),
		zap.Int("attempt", attemptCount),
	)

	// advance session
	if passed {
		if err := s.handlePass(ctx, fv); err != nil {
			return nil, err
		}
	} else {
		if err := s.handleFail(ctx, fv, "face mismatch"); err != nil {
			return nil, err
		}
	}
	return fv, nil
}

func (s *faceService) GetVerificationForUser(ctx context.Context, sessionID, userID uuid.UUID) (*models.FaceVerification, error) {
	// Ownership check if the session doesn't belong to this user,
	// GetSessionForUser returns ErrSessionNotFound.
	if _, err := s.kycSvc.GetSessionForUser(ctx, sessionID, userID); err != nil {
		return nil, err
	}

	fv, err := s.faceRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrFaceVerificationNotFound) {
			return nil, ErrFaceVerificationNotFound
		}
		return nil, ErrInternal
	}
	return fv, nil
}

// HELPERS

// handlePass advances the session to in_review so an admin can give final verdict
func (s *faceService) handlePass(ctx context.Context, fv *models.FaceVerification) error {
	if err := s.kycSvc.AdvanceStatus(ctx, fv.SessionID, models.KYCStatusInReview, StatusMeta{
		VendorName: "facepp",
	}); err != nil {
		s.logger.Error("failed to advance session after face pass",
			zap.String("session_id", fv.SessionID.String()),
			zap.Error(err),
		)
		return ErrInternal
	}
	s.logger.Info("face verification passed — session moved to in_review",
		zap.String("session_id", fv.SessionID.String()),
	)
	return nil
}

// handleFail decides whether to allow a retry or terminate the session.
func (s *faceService) handleFail(ctx context.Context, fv *models.FaceVerification, reason string) error {
	if fv.AttemptCount < maxFaceAttempts {
		// Still have retries left reset the verification to pending so the
		// user can submit another selfie. The session stays at face_verify.
		if err := s.faceRepo.UpdateResult(ctx, fv.ID, map[string]any{
			"status": models.FaceVerificationStatusPending,
		}); err != nil {
			return ErrInternal
		}
		s.logger.Info("face verification failed — retry available",
			zap.String("session_id", fv.SessionID.String()),
			zap.Int("attempt", fv.AttemptCount),
			zap.Int("max", maxFaceAttempts),
		)
		return nil
	}

	// All attempts exhausted — reject the session. The user will need to start
	// a new session after the rejection.
	s.logger.Warn("face verification failed — max attempts reached, rejecting session",
		zap.String("session_id", fv.SessionID.String()),
	)
	return s.kycSvc.AdvanceStatus(ctx, fv.SessionID, models.KYCStatusRejected, StatusMeta{
		RejectionReason: fmt.Sprintf("face verification failed after %d attempts: %s", maxFaceAttempts, reason),
	})
}

// readSelfie reads and validates the selfie from the multipart file header.
func readSelfie(fh *multipart.FileHeader) ([]byte, error) {
	if fh.Size > maxSelfieBytes {
		return nil, ErrFileTooLarge
	}

	f, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("service: open selfie file: %w", err)
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxSelfieBytes+1))
	if err != nil {
		return nil, fmt.Errorf("service: read selfie file: %w", err)
	}
	if int64(len(data)) > maxSelfieBytes {
		return nil, ErrFileTooLarge
	}
	return data, nil
}

func (s *faceService) fetchStorageBytes(ctx context.Context, bucket, key string) ([]byte, error) {
	url, err := s.storage.GetPresignedURL(ctx, bucket, key, 15*time.Minute)
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(url)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch file from storage: %s", resp.Status)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
