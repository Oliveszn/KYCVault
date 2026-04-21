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
	sessionID, ok := h.parseUUID(c, "id")
	if !ok {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	if err := c.Request.ParseMultipartForm(5 << 20); err != nil {
		respondError(c, http.StatusBadRequest, "could not parse multipart form")
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, "file field is required")
		return
	}

	fv, err := h.faceSvc.StartVerification(c.Request.Context(), services.StartVerificationRequest{
		SessionID:  sessionID,
		UserID:     userID,
		FileHeader: fileHeader,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	if err != nil {
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

	respond(c, http.StatusOK, "Face verification completed", toFaceResponse(fv))
}

// GetVerification
// GET /kyc/sessions/:id/face
// Returns the current face verification status for a session.
// The React wizard polls this after submitting the selfie.
// When status changes from "pending" to "passed" or "failed", the wizard
// either advances to the completion screen or shows the retry prompt.
func (h *FaceHandler) GetVerification(c *gin.Context) {
	sessionID, ok := h.parseUUID(c, "id")
	if !ok {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	fv, err := h.faceSvc.GetVerificationForUser(c.Request.Context(), sessionID, userID)
	if err != nil {
		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrSessionNotFound:          http.StatusNotFound,
			services.ErrFaceVerificationNotFound: http.StatusNotFound,
		})
		return
	}

	respond(c, http.StatusOK, "Face verification retrieved", toFaceResponse(fv))
}

// GetVerificationAdmin
// GET /admin/face/:id/face
// Returns the Face for a selected session for an admin
func (h *FaceHandler) GetVerificationAdmin(c *gin.Context) {
	sessionID, ok := h.parseUUID(c, "id")
	if !ok {
		return
	}

	fv, err := h.faceSvc.GetVerificationBySessionID(c.Request.Context(), sessionID)
	if err != nil {
		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrFaceVerificationNotFound: http.StatusNotFound,
		})
		return
	}

	respond(c, http.StatusOK, "face verification retrieved", toFaceResponse(fv))
}

// GetSelfieUrl
// GET /admin/face/:id/selfie-url
// Returns a presigned url for the selfie from s3
func (h *FaceHandler) GetSelfieURL(c *gin.Context) {
	verificationID, ok := h.parseUUID(c, "id")
	if !ok {
		return
	}

	url, err := h.faceSvc.GetSelfieURL(c.Request.Context(), verificationID)
	if err != nil {
		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrFaceVerificationNotFound: http.StatusNotFound,
		})
		return
	}

	respond(c, http.StatusOK, "selfie url retrieved", gin.H{
		"url":        url,
		"expires_in": "15m",
	})
}

// POST /admin/kyc/sessions/:id/face/review
// Body: { "passed": true, "note": "Face matches document." }
func (h *FaceHandler) ReviewVerification(c *gin.Context) {
	sessionID, ok := h.parseUUID(c, "id")
	if !ok {
		return
	}

	reviewerID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var dto dtos.ReviewFaceRequest
	if !h.bindJSON(c, &dto) {
		return
	}

	if err := h.faceSvc.ReviewVerification(c.Request.Context(), sessionID, reviewerID, dto.Passed, dto.Note); err != nil {
		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrFaceVerificationNotFound: http.StatusNotFound,
		})
		return
	}

	respond(c, http.StatusOK, "face verification reviewed", nil)
}

// HELPERS

func (h *FaceHandler) parseUUID(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id format")
		return uuid.Nil, false
	}
	return id, true
}

func (h *FaceHandler) bindJSON(c *gin.Context, dto any) bool {
	if err := c.ShouldBindJSON(dto); err != nil {
		h.logger.Error("bind error",
			zap.Error(err),
		)
		respondError(c, http.StatusBadRequest, err.Error())
		return false
	}
	return true
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
