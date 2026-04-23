package handlers

import (
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
	sessionID, ok := parseUUID(c, "id")
	if !ok {
		h.logger.Warn("invalid session id for upload document")
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		h.logger.Warn("unauthenticated upload document attempt")
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	h.logger.Info("upload document request received",
		zap.String("session_id", sessionID.String()),
		zap.String("user_id", userID.String()),
		zap.String("ip", c.ClientIP()),
	)

	// Parse multipart
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		h.logger.Warn("failed to parse multipart form",
			zap.Error(err),
		)
		respondError(c, http.StatusBadRequest, "could not parse multipart form")
		return
	}

	side := c.PostForm("side")
	if side == "" {
		h.logger.Warn("missing document side",
			zap.String("session_id", sessionID.String()),
			zap.String("user_id", userID.String()),
		)
		respondError(c, http.StatusBadRequest, "side field is required (front or back)")
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		h.logger.Warn("missing file in upload",
			zap.String("session_id", sessionID.String()),
			zap.String("user_id", userID.String()),
		)
		respondError(c, http.StatusBadRequest, "file field is required")
		return
	}

	h.logger.Info("processing document upload",
		zap.String("session_id", sessionID.String()),
		zap.String("user_id", userID.String()),
		zap.String("side", side),
		zap.String("filename", fileHeader.Filename),
	)

	doc, err := h.docSvc.UploadDocument(c.Request.Context(), services.UploadDocumentRequest{
		SessionID:  sessionID,
		UserID:     userID,
		Side:       side,
		FileHeader: fileHeader,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	})
	if err != nil {
		h.logger.Error("document upload failed",
			zap.String("session_id", sessionID.String()),
			zap.String("user_id", userID.String()),
			zap.String("side", side),
			zap.Error(err),
		)

		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrSessionNotFound:     http.StatusNotFound,
			services.ErrSessionWrongStage:   http.StatusConflict,
			services.ErrInvalidDocumentSide: http.StatusBadRequest,
			services.ErrInvalidFileType:     http.StatusUnsupportedMediaType,
			services.ErrFileTooLarge:        http.StatusRequestEntityTooLarge,
		})
		return
	}

	h.logger.Info("document uploaded successfully",
		zap.String("document_id", doc.ID.String()),
		zap.String("session_id", sessionID.String()),
		zap.String("user_id", userID.String()),
		zap.String("side", side),
	)

	respond(c, http.StatusCreated, "Document uploaded successfully", toDocumentResponse(doc))
}

// ListDocuments
// GET /kyc/sessions/:id/documents
// Returns all documents for a session
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	sessionID, ok := parseUUID(c, "id")
	if !ok {
		h.logger.Warn("invalid session id for list documents")
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		h.logger.Warn("unauthenticated list documents attempt")
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	h.logger.Info("list documents request",
		zap.String("session_id", sessionID.String()),
		zap.String("user_id", userID.String()),
	)

	docs, err := h.docSvc.GetDocumentsForSession(c.Request.Context(), sessionID, userID)
	if err != nil {
		h.logger.Error("failed to fetch documents",
			zap.String("session_id", sessionID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)

		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrSessionNotFound: http.StatusNotFound,
		})
		return
	}

	h.logger.Info("documents retrieved",
		zap.String("session_id", sessionID.String()),
		zap.String("user_id", userID.String()),
		zap.Int("count", len(docs)),
	)

	items := make([]documentResponse, 0, len(docs))
	for i := range docs {
		items = append(items, toDocumentResponse(&docs[i]))
	}

	respond(c, http.StatusOK, "Documents retrieved", gin.H{
		"documents": items,
		"count":     len(items),
	})
}

// GetPresignedURL
// GET /admin/documents/:doc_id/url
// Returns a 15-minute presigned URL for viewing a document image.
// Admin only never exposed to users.
func (h *DocumentHandler) GetPresignedURL(c *gin.Context) {
	docID, ok := parseUUID(c, "doc_id")
	if !ok {
		h.logger.Warn("invalid document id for presigned url")
		return
	}

	h.logger.Info("generate presigned url request",
		zap.String("doc_id", docID.String()),
	)

	url, err := h.docSvc.GetPresignedURL(c.Request.Context(), docID)
	if err != nil {
		h.logger.Error("failed to generate presigned url",
			zap.String("doc_id", docID.String()),
			zap.Error(err),
		)

		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrDocumentNotFound: http.StatusNotFound,
		})
		return
	}

	h.logger.Info("presigned url generated",
		zap.String("doc_id", docID.String()),
	)

	respond(c, http.StatusOK, "Presigned url generated", gin.H{
		"url":        url,
		"expires_in": "15m",
	})
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
