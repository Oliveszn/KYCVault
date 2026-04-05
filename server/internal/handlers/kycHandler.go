package handlers

import (
	"errors"
	"kycvault/internal/dtos"
	"kycvault/internal/middleware"
	"kycvault/internal/models"
	"kycvault/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type KYCHandler struct {
	kycSvc services.KYCService
	logger *zap.Logger
}

func NewKYCHandler(kycSvc services.KYCService, logger *zap.Logger) *KYCHandler {
	return &KYCHandler{
		kycSvc: kycSvc,
		logger: logger,
	}
}

// USERS

// InitiateSession
// POST /kyc/sessions
// Body: { "country": "NG", "id_type": "national_id" }
// Creates a new KYC session. Fails if the user already has one active.
func (h *KYCHandler) InitiateSession(c *gin.Context) {
	var dto dtos.InitiateSessionRequest
	if !h.bindJSON(c, &dto) {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	session, err := h.kycSvc.InitiateSession(c.Request.Context(), userID, dto)
	if err != nil {
		h.handleServiceError(c, err, map[error]int{
			services.ErrSessionAlreadyActive: http.StatusConflict,
		})
		return
	}

	respond(c, http.StatusCreated, "kyc session initiated", toSessionResponse(session))
}

// GetActiveSession
// GET /kyc/sessions/active
// Returns the user's current in-progress session.
// The React wizard calls this on load to resume where the user left off.
func (h *KYCHandler) GetActiveSession(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	session, err := h.kycSvc.GetActiveSessionForUser(c.Request.Context(), userID)
	if err != nil {
		h.handleServiceError(c, err, map[error]int{
			services.ErrSessionNotFound: http.StatusNotFound,
		})
		return
	}

	respond(c, http.StatusOK, "active session retrieved", toSessionResponse(session))
}

// GetSession
// GET /kyc/sessions/:id
// Returns a specific session. Ownership-checked users can only see their own.
// The React wizard polls this for status updates.
func (h *KYCHandler) GetSession(c *gin.Context) {
	sessionID, ok := h.parseUUID(c, "id")
	if !ok {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	session, err := h.kycSvc.GetSessionForUser(c.Request.Context(), sessionID, userID)
	if err != nil {
		h.handleServiceError(c, err, map[error]int{
			services.ErrSessionNotFound: http.StatusNotFound,
		})
		return
	}

	respond(c, http.StatusOK, "session retrieved", toSessionResponse(session))
}

// GetSessionHistory
// GET /kyc/sessions/history
// Returns all sessions for the user (active + terminal), newest first.
func (h *KYCHandler) GetSessionHistory(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	sessions, err := h.kycSvc.GetSessionHistoryForUser(c.Request.Context(), userID)
	if err != nil {
		h.handleServiceError(c, err, nil)
		return
	}

	items := make([]dtos.SessionResponse, 0, len(sessions))
	for i := range sessions {
		items = append(items, toSessionResponse(&sessions[i]))
	}

	respond(c, http.StatusOK, "session history retrieved", gin.H{
		"sessions": items,
		"count":    len(items),
	})
}

//ADMIN

// GetSessionQueue
// GET /admin/kyc/sessions?limit=20&offset=0
// Returns paginated sessions in IN_REVIEW status, oldest first.
func (h *KYCHandler) GetSessionQueue(c *gin.Context) {
	limit, offset := h.parsePagination(c)

	sessions, total, err := h.kycSvc.GetSessionQueue(c.Request.Context(), limit, offset)
	if err != nil {
		h.handleServiceError(c, err, nil)
		return
	}

	items := make([]dtos.SessionResponse, 0, len(sessions))
	for i := range sessions {
		items = append(items, toSessionResponse(&sessions[i]))
	}

	respond(c, http.StatusOK, "review queue retrieved", gin.H{
		"sessions": items,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetStatusCounts
// GET /admin/kyc/sessions/counts
// Returns a count per status for the dashboard tiles.
func (h *KYCHandler) GetStatusCounts(c *gin.Context) {
	counts, err := h.kycSvc.GetStatusCounts(c.Request.Context())
	if err != nil {
		h.handleServiceError(c, err, nil)
		return
	}
	respond(c, http.StatusOK, "status counts retrieved", counts)
}

// GetSessionAdmin
// GET /admin/kyc/sessions/:id
// Admin view of a session — no ownership check.
func (h *KYCHandler) GetSessionAdmin(c *gin.Context) {
	sessionID, ok := h.parseUUID(c, "id")
	if !ok {
		return
	}

	session, err := h.kycSvc.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		h.handleServiceError(c, err, map[error]int{
			services.ErrSessionNotFound: http.StatusNotFound,
		})
		return
	}

	respond(c, http.StatusOK, "session retrieved", toSessionResponse(session))
}

// ApproveSession
// POST /admin/kyc/sessions/:id/approve
// Body: { "note": "All documents verified manually." }
func (h *KYCHandler) ApproveSession(c *gin.Context) {
	sessionID, ok := h.parseUUID(c, "id")
	if !ok {
		return
	}

	var dto dtos.ReviewSessionRequest
	if !h.bindJSON(c, &dto) {
		return
	}

	reviewerID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	if err := h.kycSvc.ApproveSession(c.Request.Context(), sessionID, reviewerID, dto.Note); err != nil {
		h.handleServiceError(c, err, map[error]int{
			services.ErrSessionNotFound:         http.StatusNotFound,
			services.ErrSessionAlreadyTerminal:  http.StatusConflict,
			services.ErrInvalidStatusTransition: http.StatusUnprocessableEntity,
		})
		return
	}

	respond(c, http.StatusOK, "session approved", nil)
}

// RejectSession
// POST /admin/kyc/sessions/:id/reject
// Body: { "note": "Internal note.", "reason": "Document image was blurry." }
func (h *KYCHandler) RejectSession(c *gin.Context) {
	sessionID, ok := h.parseUUID(c, "id")
	if !ok {
		return
	}

	var dto dtos.RejectSessionRequest
	if !h.bindJSON(c, &dto) {
		return
	}

	reviewerID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	if err := h.kycSvc.RejectSession(c.Request.Context(), sessionID, reviewerID, dto.Note, dto.Reason); err != nil {
		h.handleServiceError(c, err, map[error]int{
			services.ErrSessionNotFound:         http.StatusNotFound,
			services.ErrSessionAlreadyTerminal:  http.StatusConflict,
			services.ErrInvalidStatusTransition: http.StatusUnprocessableEntity,
		})
		return
	}

	respond(c, http.StatusOK, "session rejected", nil)
}

// helpers

func (h *KYCHandler) bindJSON(c *gin.Context, dto any) bool {
	if err := c.ShouldBindJSON(dto); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}

func (h *KYCHandler) parseUUID(c *gin.Context, param string) (uuid.UUID, bool) {
	raw := c.Param(param)
	id, err := uuid.Parse(raw)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id format")
		return uuid.Nil, false
	}
	return id, true
}

func (h *KYCHandler) parsePagination(c *gin.Context) (limit, offset int) {
	limit = 20
	offset = 0
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}
	return
}

func (h *KYCHandler) handleServiceError(c *gin.Context, err error, statusMap map[error]int) {
	for target, code := range statusMap {
		if errors.Is(err, target) {
			respondError(c, code, err.Error())
			return
		}
	}
	h.logger.Error("unhandled service error", zap.Error(err))
	respondError(c, http.StatusInternalServerError, "an internal error occurred")
}

// toSessionResponse converts the model to the API response shape.
// Sensitive fields (vendor_raw_result, storage keys) are never included.
func toSessionResponse(s *models.KYCSession) dtos.SessionResponse {
	resp := dtos.SessionResponse{
		ID:            s.ID,
		Status:        string(s.Status),
		Country:       s.Country,
		IDType:        string(s.IDType),
		AttemptNumber: s.AttemptNumber,
		CreatedAt:     s.CreatedAt,
		UpdatedAt:     s.UpdatedAt,
	}

	if s.RejectionReason != "" {
		resp.RejectionReason = s.RejectionReason
	}
	if s.ExpiresAt != nil {
		resp.ExpiresAt = s.ExpiresAt
	}

	// Map documents if preloaded
	if len(s.Documents) > 0 {
		docs := make([]dtos.DocumentSummary, 0, len(s.Documents))
		for _, d := range s.Documents {
			docs = append(docs, dtos.DocumentSummary{
				ID:     d.ID,
				Side:   string(d.Side),
				Status: string(d.Status),
			})
		}
		resp.Documents = docs
	}

	return resp
}
