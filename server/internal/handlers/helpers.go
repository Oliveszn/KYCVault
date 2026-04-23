package handlers

import (
	"errors"
	"kycvault/internal/dtos"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func respond(c *gin.Context, status int, message string, payload interface{}) {
	c.JSON(status, dtos.StructuredResponse{
		Success: status < 400,
		Status:  status,
		Message: message,
		Payload: payload,
	})
}

func respondError(c *gin.Context, status int, message string) {
	respond(c, status, message, nil)
}

func handleServiceError(c *gin.Context, logger *zap.Logger, err error, statusMap map[error]int) {
	for target, code := range statusMap {
		if errors.Is(err, target) {
			respondError(c, code, err.Error())
			return
		}
	}
	if logger != nil {
		logger.Error("unhandled service error", zap.Error(err))
	}
	respondError(c, http.StatusInternalServerError, "an internal error occurred")
}

// bindJSON attempts to bind and validate the request body. On failure it
// writes the error response and returns false so the caller can return early.
func bindJSON(c *gin.Context, logger *zap.Logger, dto interface{}) bool {
	if err := c.ShouldBindJSON(dto); err != nil {
		logger.Error("bind error",
			zap.Error(err),
		)
		respondError(c, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid id format")
		return uuid.Nil, false
	}
	return id, true
}
