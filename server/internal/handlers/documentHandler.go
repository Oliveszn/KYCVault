package handlers

import (
	"errors"
	"kycvault/internal/middleware"
	"kycvault/internal/models"
	"kycvault/internal/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type DocumentHandler struct {
	docSvc services.DocumentService
	logger *zap.Logger
}

func NewDocumentHandler(docSvc services.DocumentService, logger *zap.Logger) *DocumentHandler {
	return &DocumentHandler{
		docSvc: docSvc,
		logger: logger,
	}
}

// UploadDocument
// POST /kyc/sessions/:id/documents
// Content-Type: multipart/form-data
// Fields:
//
//	side  — "front" or "back"
//	file  — the image or PDF
//
// This handler does only two things: parse the multipart form and hand
// everything to the service. Zero business logic lives here.
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	sessionID, ok := h.parseUUID(c, "id")
	if !ok {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	// Parse multipart — 10 MB limit enforced by Gin before we even touch the file.
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		respondError(c, http.StatusBadRequest, "could not parse multipart form")
		return
	}

	side := c.PostForm("side")
	if side == "" {
		respondError(c, http.StatusBadRequest, "side field is required (front or back)")
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		respondError(c, http.StatusBadRequest, "file field is required")
		return
	}

	doc, err := h.docSvc.UploadDocument(c.Request.Context(), services.UploadDocumentRequest{
		SessionID:  sessionID,
		UserID:     userID,
		Side:       side,
		FileHeader: fileHeader,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	if err != nil {
		h.handleServiceError(c, err, map[error]int{
			services.ErrSessionNotFound:     http.StatusNotFound,
			services.ErrSessionWrongStage:   http.StatusConflict,
			services.ErrInvalidDocumentSide: http.StatusBadRequest,
			services.ErrInvalidFileType:     http.StatusUnsupportedMediaType,
			services.ErrFileTooLarge:        http.StatusRequestEntityTooLarge,
		})
		return
	}

	respond(c, http.StatusCreated, "document uploaded successfully", toDocumentResponse(doc))
}

// ListDocuments
// GET /kyc/sessions/:id/documents
// Returns all documents for a session
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	sessionID, ok := h.parseUUID(c, "id")
	if !ok {
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	docs, err := h.docSvc.GetDocumentsForSession(c.Request.Context(), sessionID, userID)
	if err != nil {
		h.handleServiceError(c, err, map[error]int{
			services.ErrSessionNotFound: http.StatusNotFound,
		})
		return
	}

	items := make([]documentResponse, 0, len(docs))
	for i := range docs {
		items = append(items, toDocumentResponse(&docs[i]))
	}

	respond(c, http.StatusOK, "documents retrieved", gin.H{
		"documents": items,
		"count":     len(items),
	})
}

// GetPresignedURL
// GET /admin/documents/:doc_id/url
// Returns a 15-minute presigned URL for viewing a document image.
// Admin only never exposed to users.
func (h *DocumentHandler) GetPresignedURL(c *gin.Context) {
	docID, ok := h.parseUUID(c, "doc_id")
	if !ok {
		return
	}

	url, err := h.docSvc.GetPresignedURL(c.Request.Context(), docID)
	if err != nil {
		h.handleServiceError(c, err, map[error]int{
			services.ErrDocumentNotFound: http.StatusNotFound,
		})
		return
	}

	respond(c, http.StatusOK, "presigned url generated", gin.H{
		"url":        url,
		"expires_in": "15m",
	})
}

// HELPERS

func (h *DocumentHandler) parseUUID(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id format")
		return uuid.Nil, false
	}
	return id, true
}

func (h *DocumentHandler) handleServiceError(c *gin.Context, err error, statusMap map[error]int) {
	for target, code := range statusMap {
		if errors.Is(err, target) {
			respondError(c, code, err.Error())
			return
		}
	}
	h.logger.Error("unhandled document service error", zap.Error(err))
	respondError(c, http.StatusInternalServerError, "an internal error occurred")
}

// Response shape

// documentResponse is what the API returns. Storage keys, checksums, and
// extracted OCR data are deliberately excluded from all user-facing responses.
type documentResponse struct {
	ID               uuid.UUID `json:"id"`
	SessionID        uuid.UUID `json:"session_id"`
	Side             string    `json:"side"`
	Status           string    `json:"status"`
	MIMEType         string    `json:"mime_type"`
	FileSizeBytes    int64     `json:"file_size_bytes"`
	OriginalFilename string    `json:"original_filename"`
	RejectionReason  string    `json:"rejection_reason,omitempty"`
	UploadedAt       time.Time `json:"uploaded_at"`
}

func toDocumentResponse(doc *models.KYCDocument) documentResponse {
	return documentResponse{
		ID:               doc.ID,
		SessionID:        doc.SessionID,
		Side:             string(doc.Side),
		Status:           string(doc.Status),
		MIMEType:         doc.MIMEType,
		FileSizeBytes:    doc.FileSizeBytes,
		OriginalFilename: doc.OriginalFilename,
		RejectionReason:  doc.RejectionReason,
		UploadedAt:       doc.UploadedAt,
	}
}
