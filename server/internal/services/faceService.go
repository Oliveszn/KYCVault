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
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrFaceVerificationNotFound = errors.New("Face verification not found")
	ErrFaceSessionWrongStage    = errors.New("Session is not at the face verification stage")
	ErrFaceMaxAttemptsReached   = errors.New("Maximum face verification attempts reached")
	ErrFaceDocumentNotReady     = errors.New("Document images are not available for face matching")
	ErrFaceUnsupportedIDType    = errors.New("ID type is not supported for this country")
	ErrFaceVerificationPending  = errors.New("Face verification is already in progress")
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

	//FOR ADMINS
	GetVerificationBySessionID(ctx context.Context, sessionID uuid.UUID) (*models.FaceVerification, error)
	GetSelfieURL(ctx context.Context, verificationID uuid.UUID) (string, error)
	ReviewVerification(ctx context.Context, verificationID, reviewerID uuid.UUID, passed bool, note string) error
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
	s.logger.Info("starting face verification",
		zap.String("session_id", req.SessionID.String()),
		zap.String("user_id", req.UserID.String()),
	)

	session, err := s.kycSvc.GetSessionForUser(ctx, req.SessionID, req.UserID)
	if err != nil {
		s.logger.Warn("failed to fetch session for face verification",
			zap.String("session_id", req.SessionID.String()),
			zap.String("user_id", req.UserID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	if session.Status != models.KYCStatusFaceVerify {
		s.logger.Warn("face verification attempted in wrong session stage",
			zap.String("session_id", req.SessionID.String()),
			zap.String("current_status", string(session.Status)),
		)
		return nil, ErrFaceSessionWrongStage
	}

	existing, err := s.faceRepo.GetBySessionID(ctx, req.SessionID)
	if err != nil && !errors.Is(err, repository.ErrFaceVerificationNotFound) {
		s.logger.Error("failed to fetch existing face verification",
			zap.String("session_id", req.SessionID.String()),
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	if existing != nil && existing.Status == models.FaceVerificationStatusPending {
		s.logger.Warn("face verification already pending",
			zap.String("session_id", req.SessionID.String()),
		)
		return nil, ErrFaceVerificationPending
	}

	if existing != nil && existing.AttemptCount >= maxFaceAttempts {
		s.logger.Warn("max face verification attempts reached",
			zap.String("session_id", req.SessionID.String()),
			zap.Int("attempts", existing.AttemptCount),
		)
		return nil, ErrFaceMaxAttemptsReached
	}

	selfieBytes, err := readSelfie(req.FileHeader)
	if err != nil {
		s.logger.Warn("invalid selfie upload",
			zap.String("session_id", req.SessionID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	checksum := computeChecksum(selfieBytes)

	verificationID := uuid.New()
	if existing != nil {
		verificationID = existing.ID
	}

	attemptCount := 1
	if existing != nil {
		attemptCount = existing.AttemptCount + 1
	}

	selfieKey := storage.KeyForSelfie(req.SessionID.String(), verificationID.String())

	fv := &models.FaceVerification{
		ID:                  verificationID,
		SessionID:           req.SessionID,
		UserID:              req.UserID,
		Status:              models.FaceVerificationStatusPending,
		SelfieStorageKey:    selfieKey,
		SelfieStorageBucket: s.bucket,
		SelfieChecksum:      checksum,
		VendorName:          "manual",
		VendorRequestID:     verificationID.String(),
		AttemptCount:        attemptCount,
	}

	if err := s.faceRepo.UpsertVerification(ctx, fv); err != nil {
		s.logger.Error("failed to upsert face verification record",
			zap.String("session_id", req.SessionID.String()),
			zap.String("verification_id", verificationID.String()),
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	if err := s.storage.Put(
		ctx,
		s.bucket,
		selfieKey,
		bytes.NewReader(selfieBytes),
		int64(len(selfieBytes)),
		"image/jpeg",
	); err != nil {
		s.logger.Error("failed to upload selfie to storage",
			zap.String("verification_id", verificationID.String()),
			zap.String("key", selfieKey),
			zap.Error(err),
		)

		_ = s.faceRepo.UpdateResult(ctx, verificationID, map[string]any{
			"status":         models.FaceVerificationStatusFailed,
			"failure_reason": "selfie upload failed; please retry",
		})

		return nil, ErrInternal
	}

	if err := s.kycSvc.AdvanceStatus(
		ctx,
		req.SessionID,
		models.KYCStatusInReview,
		StatusMeta{
			VendorName:      "manual",
			VendorSessionID: verificationID.String(),
		},
	); err != nil {
		s.logger.Error("failed to advance session to in_review",
			zap.String("session_id", req.SessionID.String()),
			zap.Error(err),
		)
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
	})

	s.logger.Info("face verification submitted successfully",
		zap.String("verification_id", verificationID.String()),
		zap.String("session_id", req.SessionID.String()),
		zap.Int("attempt_count", attemptCount),
	)

	return fv, nil
}

func (s *faceService) GetVerificationForUser(ctx context.Context, sessionID, userID uuid.UUID) (*models.FaceVerification, error) {
	s.logger.Info("fetching face verification for user",
		zap.String("session_id", sessionID.String()),
		zap.String("user_id", userID.String()),
	)

	if _, err := s.kycSvc.GetSessionForUser(ctx, sessionID, userID); err != nil {
		s.logger.Warn("unauthorized access to face verification",
			zap.String("session_id", sessionID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	fv, err := s.faceRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrFaceVerificationNotFound) {
			return nil, ErrFaceVerificationNotFound
		}
		s.logger.Error("failed to fetch face verification",
			zap.String("session_id", sessionID.String()),
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	return fv, nil
}

// ADMIN
func (s *faceService) GetVerificationBySessionID(ctx context.Context, sessionID uuid.UUID) (*models.FaceVerification, error) {
	s.logger.Info("admin fetching face verification",
		zap.String("session_id", sessionID.String()),
	)

	fv, err := s.faceRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrFaceVerificationNotFound) {
			return nil, ErrFaceVerificationNotFound
		}
		s.logger.Error("failed to fetch face verification (admin)",
			zap.String("session_id", sessionID.String()),
			zap.Error(err),
		)
		return nil, ErrInternal
	}
	return fv, nil
}

func (s *faceService) GetSelfieURL(ctx context.Context, verificationID uuid.UUID) (string, error) {
	s.logger.Info("generating selfie presigned url",
		zap.String("verification_id", verificationID.String()),
	)

	fv, err := s.faceRepo.GetByID(ctx, verificationID)
	if err != nil {
		s.logger.Warn("face verification not found for selfie url",
			zap.String("verification_id", verificationID.String()),
		)
		return "", ErrFaceVerificationNotFound
	}

	url, err := s.storage.GetPresignedURL(ctx, fv.SelfieStorageBucket, fv.SelfieStorageKey, 15*time.Minute)
	if err != nil {
		s.logger.Error("failed to generate presigned url",
			zap.String("verification_id", verificationID.String()),
			zap.Error(err),
		)
		return "", ErrInternal
	}

	return url, nil
}

func (s *faceService) ReviewVerification(ctx context.Context, verificationID, reviewerID uuid.UUID, passed bool, note string) error {
	s.logger.Info("reviewing face verification",
		zap.String("verification_id", verificationID.String()),
		zap.String("reviewer_id", reviewerID.String()),
		zap.Bool("passed", passed),
	)

	fv, err := s.faceRepo.GetByID(ctx, verificationID)
	if err != nil {
		s.logger.Warn("face verification not found during review",
			zap.String("verification_id", verificationID.String()),
		)
		return ErrFaceVerificationNotFound
	}

	status := models.FaceVerificationStatusFailed
	if passed {
		status = models.FaceVerificationStatusPassed
	}

	if err := s.faceRepo.UpdateResult(ctx, verificationID, map[string]any{
		"status":         status,
		"failure_reason": note,
		"vendor_name":    "manual",
	}); err != nil {
		s.logger.Error("failed to update face verification result",
			zap.String("verification_id", verificationID.String()),
			zap.Error(err),
		)
		return ErrInternal
	}

	if passed {
		if err := s.handlePass(ctx, fv); err != nil {
			s.logger.Error("failed to handle passed verification",
				zap.String("verification_id", verificationID.String()),
				zap.Error(err),
			)
			return err
		}
	} else {
		if err := s.handleFail(ctx, fv, note); err != nil {
			s.logger.Error("failed to handle failed verification",
				zap.String("verification_id", verificationID.String()),
				zap.Error(err),
			)
			return err
		}
	}

	s.audit.Log(ctx, models.AuditEvent{
		ActorID:   &reviewerID,
		ActorRole: "admin",
		SessionID: &fv.SessionID,
		UserID:    &fv.UserID,
		EventType: models.AuditEventFaceVerifyStarted,
	})

	s.logger.Info("face verification review completed",
		zap.String("verification_id", verificationID.String()),
		zap.Bool("passed", passed),
	)

	return nil
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

	// 	// All attempts exhausted — reject the session. The user will need to start
	// 	// a new session after the rejection.
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
