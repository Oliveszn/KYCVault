package handlers

import (
	"kycvault/internal/dtos"
	"kycvault/internal/middleware"
	"kycvault/internal/models"
	"kycvault/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type FaceHandler struct {
	faceSvc services.FaceService
	logger  *zap.Logger
}

func NewFaceHandler(
	faceSvc services.FaceService,

	logger *zap.Logger,
) *FaceHandler {
	return &FaceHandler{
		faceSvc: faceSvc,

		logger: logger,
	}
}

// StartVerification
// POST /kyc/sessions/:id/face
// Content-Type: multipart/form-data
// Fields:
//
//	file — the selfie image (JPEG, max 5 MB)
func (h *FaceHandler) StartVerification(c *gin.Context) {
	sessionID, ok := parseUUID(c, "id")
	if !ok {
		h.logger.Warn("invalid session id for face verification")
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		h.logger.Warn("unauthenticated face verification attempt")
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	h.logger.Info("face verification request received",
		zap.String("session_id", sessionID.String()),
		zap.String("user_id", userID.String()),
		zap.String("ip", c.ClientIP()),
	)

	if err := c.Request.ParseMultipartForm(5 << 20); err != nil {
		h.logger.Warn("failed to parse multipart form for face verification",
			zap.Error(err),
		)
		respondError(c, http.StatusBadRequest, "could not parse multipart form")
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		h.logger.Warn("missing selfie file",
			zap.String("session_id", sessionID.String()),
			zap.String("user_id", userID.String()),
		)
		respondError(c, http.StatusBadRequest, "file field is required")
		return
	}

	h.logger.Info("processing face verification",
		zap.String("session_id", sessionID.String()),
		zap.String("user_id", userID.String()),
		zap.String("filename", fileHeader.Filename),
	)

	fv, err := h.faceSvc.StartVerification(c.Request.Context(), services.StartVerificationRequest{
		SessionID:  sessionID,
		UserID:     userID,
		FileHeader: fileHeader,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	if err != nil {
		h.logger.Error("face verification failed",
			zap.String("session_id", sessionID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)

		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrSessionNotFound:         http.StatusNotFound,
			services.ErrFaceSessionWrongStage:   http.StatusConflict,
			services.ErrFaceVerificationPending: http.StatusConflict,
			services.ErrFaceMaxAttemptsReached:  http.StatusUnprocessableEntity,
			services.ErrFaceDocumentNotReady:    http.StatusUnprocessableEntity,
			services.ErrFileTooLarge:            http.StatusRequestEntityTooLarge,
		})
		return
	}

	h.logger.Info("face verification completed",
		zap.String("verification_id", fv.ID.String()),
		zap.String("session_id", sessionID.String()),
		zap.String("user_id", userID.String()),
		zap.String("status", string(fv.Status)),
	)

	respond(c, http.StatusOK, "Face verification completed", toFaceResponse(fv))
}

// GetVerification
// GET /kyc/sessions/:id/face
// Returns the current face verification status for a session.
// The React wizard polls this after submitting the selfie.
// When status changes from "pending" to "passed" or "failed", the wizard
// either advances to the completion screen or shows the retry prompt.
func (h *FaceHandler) GetVerification(c *gin.Context) {
	sessionID, ok := parseUUID(c, "id")
	if !ok {
		h.logger.Warn("invalid session id for get verification")
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		h.logger.Warn("unauthenticated get verification attempt")
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	h.logger.Info("fetching face verification",
		zap.String("session_id", sessionID.String()),
		zap.String("user_id", userID.String()),
	)

	fv, err := h.faceSvc.GetVerificationForUser(c.Request.Context(), sessionID, userID)
	if err != nil {
		h.logger.Error("failed to fetch face verification",
			zap.String("session_id", sessionID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)

		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrSessionNotFound:          http.StatusNotFound,
			services.ErrFaceVerificationNotFound: http.StatusNotFound,
		})
		return
	}

	h.logger.Info("face verification retrieved",
		zap.String("verification_id", fv.ID.String()),
		zap.String("status", string(fv.Status)),
	)

	respond(c, http.StatusOK, "Face verification retrieved", toFaceResponse(fv))
}

// GetVerificationAdmin
// GET /admin/face/:id/face
// Returns the Face for a selected session for an admin
func (h *FaceHandler) GetVerificationAdmin(c *gin.Context) {
	sessionID, ok := parseUUID(c, "id")
	if !ok {
		h.logger.Warn("invalid session id for admin face fetch")
		return
	}

	h.logger.Info("admin fetching face verification",
		zap.String("session_id", sessionID.String()),
	)

	fv, err := h.faceSvc.GetVerificationBySessionID(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("failed to fetch face verification (admin)",
			zap.String("session_id", sessionID.String()),
			zap.Error(err),
		)

		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrFaceVerificationNotFound: http.StatusNotFound,
		})
		return
	}

	h.logger.Info("face verification retrieved (admin)",
		zap.String("verification_id", fv.ID.String()),
	)

	respond(c, http.StatusOK, "face verification retrieved", toFaceResponse(fv))
}

// GetSelfieUrl
// GET /admin/face/:id/selfie-url
// Returns a presigned url for the selfie from s3
func (h *FaceHandler) GetSelfieURL(c *gin.Context) {
	verificationID, ok := parseUUID(c, "id")
	if !ok {
		h.logger.Warn("invalid verification id for selfie url")
		return
	}

	h.logger.Info("generate selfie url request",
		zap.String("verification_id", verificationID.String()),
	)

	url, err := h.faceSvc.GetSelfieURL(c.Request.Context(), verificationID)
	if err != nil {
		h.logger.Error("failed to generate selfie url",
			zap.String("verification_id", verificationID.String()),
			zap.Error(err),
		)

		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrFaceVerificationNotFound: http.StatusNotFound,
		})
		return
	}

	h.logger.Info("selfie url generated",
		zap.String("verification_id", verificationID.String()),
	)

	respond(c, http.StatusOK, "selfie url retrieved", gin.H{
		"url":        url,
		"expires_in": "15m",
	})
}

// POST /admin/kyc/sessions/:id/face/review
// Body: { "passed": true, "note": "Face matches document." }
func (h *FaceHandler) ReviewVerification(c *gin.Context) {
	sessionID, ok := parseUUID(c, "id")
	if !ok {
		h.logger.Warn("invalid session id for face review")
		return
	}

	reviewerID, ok := middleware.GetUserID(c)
	if !ok {
		h.logger.Warn("unauthenticated face review attempt")
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var dto dtos.ReviewFaceRequest
	if !bindJSON(c, h.logger, &dto) {
		return
	}

	h.logger.Info("admin reviewing face verification",
		zap.String("session_id", sessionID.String()),
		zap.String("reviewer_id", reviewerID.String()),
		zap.Bool("passed", dto.Passed),
	)

	if err := h.faceSvc.ReviewVerification(
		c.Request.Context(),
		sessionID,
		reviewerID,
		dto.Passed,
		dto.Note,
	); err != nil {

		h.logger.Error("face review failed",
			zap.String("session_id", sessionID.String()),
			zap.String("reviewer_id", reviewerID.String()),
			zap.Error(err),
		)

		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrFaceVerificationNotFound: http.StatusNotFound,
		})
		return
	}

	h.logger.Info("face verification reviewed successfully",
		zap.String("session_id", sessionID.String()),
		zap.String("reviewer_id", reviewerID.String()),
		zap.Bool("passed", dto.Passed),
	)

	respond(c, http.StatusOK, "face verification reviewed", nil)
}

//Response shape

// faceResponse is the user-facing shape. Raw scores, storage keys, and
// vendor payloads are never returned. We return enough for the React wizard
// to know what state the user is in and what to display.
type faceResponse struct {
	ID             uuid.UUID `json:"id"`
	SessionID      uuid.UUID `json:"session_id"`
	Status         string    `json:"status"`
	LivenessPassed bool      `json:"liveness_passed"`
	MatchPassed    bool      `json:"match_passed"`
	AttemptCount   int       `json:"attempt_count"`
	AttemptsLeft   int       `json:"attempts_left"`
	FailureReason  string    `json:"failure_reason,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func toFaceResponse(fv *models.FaceVerification) faceResponse {
	attemptsLeft := 3 - fv.AttemptCount
	if attemptsLeft < 0 {
		attemptsLeft = 0
	}
	return faceResponse{
		ID:             fv.ID,
		SessionID:      fv.SessionID,
		Status:         string(fv.Status),
		LivenessPassed: fv.LivenessPassed,
		MatchPassed:    fv.MatchPassed,
		AttemptCount:   fv.AttemptCount,
		AttemptsLeft:   attemptsLeft,
		FailureReason:  fv.FailureReason,
		CreatedAt:      fv.CreatedAt,
		UpdatedAt:      fv.UpdatedAt,
	}
}
