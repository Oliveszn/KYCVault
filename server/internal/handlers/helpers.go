package handlers

import (
	"errors"
	"kycvault/internal/dtos"
	"net/http"

	"github.com/gin-gonic/gin"
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
