package handlers

import (
	"kycvault/internal/dtos"

	"github.com/gin-gonic/gin"
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
