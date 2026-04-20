package middleware

import (
	"errors"
	"kycvault/internal/dtos"
	"kycvault/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	//handlers receive identity via these context keys
	ContextKeyUserID = "authed_user_id"
	ContextKeyRole   = "authed_user_role"
)

// Authenticate is a middleware that validates the access token on request, rejects expired or missing token
func Authenticate(jwtUtil *utils.JWTUtil, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractBearerToken(c)
		if err != nil {
			logger.Warn("missing or malformed Authorization header",
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.FullPath()),
			)
			abortWithError(c, http.StatusUnauthorized, "authorization header must be: Bearer <token>")
			return
		}

		claims, err := jwtUtil.ValidateAccessToken(token)
		if err != nil {
			status := http.StatusUnauthorized
			msg := "invalid token"

			switch {
			case errors.Is(err, utils.ErrTokenExpired):
				msg = "access token has expired"
			case errors.Is(err, utils.ErrTokenMalformed):
				msg = "malformed token"
				logger.Warn("malformed token received",
					zap.String("ip", c.ClientIP()),
				)
			default:
				logger.Warn("token validation failed",
					zap.String("error", err.Error()),
					zap.String("ip", c.ClientIP()),
				)
			}

			abortWithError(c, status, msg)
			return
		}

		// Stamp validated identity onto the context. Downstream handlers and
		// middleware read from here — they never touch the token again.
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyRole, claims.Role)

		c.Next()
	}
}

// Requirerole enforces a minimum role, comes after authenticate
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	roleSet := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[r] = struct{}{}
	}

	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			//added this error incase require role is used before or without authenticate
			abortWithError(c, http.StatusInternalServerError, "identity not established")
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			abortWithError(c, http.StatusInternalServerError, "malformed identity context")
			return
		}

		if _, allowed := roleSet[roleStr]; !allowed {
			abortWithError(c, http.StatusForbidden, "you do not have permission to access this resource")
			return
		}

		c.Next()
	}
}

// GetUserId gets the user's id
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	v, exists := c.Get(ContextKeyUserID)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

// GetUserRole gets the user's role
func GetUserRole(c *gin.Context) (string, bool) {
	v, exists := c.Get(ContextKeyRole)
	if !exists {
		return "", false
	}
	role, ok := v.(string)
	return role, ok
}

// extractBeraretoken  pulls the token from "Bearer token"
// func extractBearerToken(c *gin.Context) (string, error) {
// 	header := c.GetHeader("Authorization")
// 	if header == "" {
// 		return "", errors.New("authorization header is missing")
// 	}

// 	parts := strings.SplitN(header, " ", 2)
// 	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
// 		return "", errors.New("authorization header format must be: Bearer <token>")
// 	}

// 	token := strings.TrimSpace(parts[1])
// 	if token == "" {
// 		return "", errors.New("bearer token is empty")
// 	}

// 	return token, nil
// }

func extractBearerToken(c *gin.Context) (string, error) {
	// try Authorization header first, normal HTTP
	header := c.GetHeader("Authorization")
	if header != "" {
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return "", errors.New("authorization header format must be: Bearer <token>")
		}
		token := strings.TrimSpace(parts[1])
		if token == "" {
			return "", errors.New("bearer token is empty")
		}
		return token, nil
	}

	// fallback to cookie for WebSocket upgrades
	token, err := c.Cookie("access_token")
	if err == nil && strings.TrimSpace(token) != "" {
		return token, nil
	}

	return "", errors.New("authorization header is missing")
}

// abortWithError writes a structured error response and halts the middleware chain.
func abortWithError(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, dtos.StructuredResponse{
		Success: false,
		Status:  status,
		Message: message,
		Payload: nil,
	})
}
