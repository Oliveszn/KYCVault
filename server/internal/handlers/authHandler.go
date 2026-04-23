package handlers

import (
	"errors"
	"kycvault/internal/dtos"
	"kycvault/internal/middleware"
	"kycvault/internal/services"
	"kycvault/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authSvc   services.AuthService
	jwtUtil   *utils.JWTUtil
	cookieCfg utils.CookieConfig
	logger    *zap.Logger
}

// NewAuthHandler constructs and returns an AuthHandler with all deps injected.
func NewAuthHandler(
	authSvc services.AuthService,
	jwtUtil *utils.JWTUtil,
	cookieCfg utils.CookieConfig,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		authSvc:   authSvc,
		jwtUtil:   jwtUtil,
		cookieCfg: cookieCfg,
		logger:    logger,
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account with the provided credentials.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dtos.RegisterUserDto true "Registration payload"
// @Success      201  {object}  dtos.StructuredResponse
// @Failure      400  {object}  dtos.StructuredResponse
// @Failure      409  {object}  dtos.StructuredResponse
// @Failure      500  {object}  dtos.StructuredResponse
// @Router       /auth/register [post]
// func (h *AuthHandler) Register(c *gin.Context) {
// 	var dto dtos.RegisterUserDto
// 	if !h.bindJSON(c, &dto) {
// 		return
// 	}

// 	user, err := h.authSvc.Register(c.Request.Context(), dto)
// 	if err != nil {
// 		h.handleServiceError(c, err, map[error]int{
// 			services.ErrUserAlreadyExists: http.StatusConflict,
// 			services.ErrPasswordMismatch:  http.StatusBadRequest,
// 		})
// 		return
// 	}

//		respond(c, http.StatusCreated, "account created successfully", gin.H{
//			"id":        user.ID,
//			"email":     user.Email,
//			"firstName": user.FirstName,
//			"lastName":  user.LastName,
//			"role":      user.Role,
//		})
//	}
func (h *AuthHandler) Register(c *gin.Context) {
	var dto dtos.RegisterUserDto
	if !bindJSON(c, h.logger, &dto) {
		h.logger.Warn("invalid register payload", zap.String("ip", c.ClientIP()))
		return
	}

	h.logger.Info("register attempt",
		zap.String("email", dto.Email),
		zap.String("ip", c.ClientIP()),
	)

	pair, err := h.authSvc.Register(
		c.Request.Context(),
		dto,
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		h.logger.Warn("register failed",
			zap.String("email", dto.Email),
			zap.Error(err),
		)
		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrUserAlreadyExists: http.StatusConflict,
			services.ErrPasswordMismatch:  http.StatusBadRequest,
		})
		return
	}

	h.logger.Info("user registered successfully",
		zap.String("email", dto.Email),
	)

	//set refresh token cookie
	utils.SetRefreshTokenCookie(
		c,
		pair.RawRefreshToken,
		h.jwtUtil.RefreshTokenTTL(),
		h.cookieCfg,
	)

	respond(c, http.StatusCreated, "Account created successfully", gin.H{
		"accessToken": pair.AccessToken,
		"expiresIn":   pair.ExpiresIn,
		"tokenType":   "Bearer",
	})
}

// Login godoc
// @Summary      Authenticate a user
// @Description  Validates credentials and returns an access token + sets a refresh token cookie.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body dtos.LoginUserDto true "Login payload"
// @Success      200  {object}  dtos.StructuredResponse
// @Failure      400  {object}  dtos.StructuredResponse
// @Failure      401  {object}  dtos.StructuredResponse
// @Failure      500  {object}  dtos.StructuredResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var dto dtos.LoginUserDto
	if !bindJSON(c, h.logger, &dto) {
		h.logger.Warn("invalid login payload", zap.String("ip", c.ClientIP()))
		return
	}

	h.logger.Info("login attempt",
		zap.String("email", dto.Email),
		zap.String("ip", c.ClientIP()),
	)

	pair, err := h.authSvc.Login(
		c.Request.Context(),
		dto,
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		h.logger.Warn("login failed",
			zap.String("email", dto.Email),
			zap.Error(err),
		)
		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrInvalidCredentials: http.StatusUnauthorized,
		})
		return
	}

	h.logger.Info("login successful",
		zap.String("email", dto.Email),
	)

	// Refresh token goes into a secure httpOnly cookie — never in the body.
	utils.SetRefreshTokenCookie(c, pair.RawRefreshToken, h.jwtUtil.RefreshTokenTTL(), h.cookieCfg)

	respond(c, http.StatusOK, "Login successful", gin.H{
		"accessToken": pair.AccessToken,
		"expiresIn":   pair.ExpiresIn,
		"tokenType":   "Bearer",
	})
}

// Refresh godoc
// @Summary      Rotate tokens
// @Description  Validates the refresh token cookie, rotates it, and issues a new access token.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  dtos.StructuredResponse
// @Failure      401  {object}  dtos.StructuredResponse
// @Failure      500  {object}  dtos.StructuredResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	rawToken, err := c.Cookie(utils.RefreshTokenCookieName)
	if err != nil || rawToken == "" {
		h.logger.Warn("missing refresh token", zap.String("ip", c.ClientIP()))
		respondError(c, http.StatusUnauthorized, "refresh token cookie is missing or invalid")
		return
	}

	h.logger.Info("refresh attempt", zap.String("ip", c.ClientIP()))

	pair, err := h.authSvc.RefreshTokens(
		c.Request.Context(),
		rawToken,
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		// On reuse detection, also clear the cookie so the client is fully signed out.
		if errors.Is(err, services.ErrRefreshTokenReuse) {
			h.logger.Warn("refresh token reuse detected", zap.String("ip", c.ClientIP()))
			utils.ClearRefreshTokenCookie(c, h.cookieCfg)
		}
		h.logger.Warn("refresh failed", zap.Error(err))
		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrRefreshTokenInvalid: http.StatusUnauthorized,
			services.ErrRefreshTokenExpired: http.StatusUnauthorized,
			services.ErrRefreshTokenReuse:   http.StatusUnauthorized,
		})
		return
	}

	h.logger.Info("tokens refreshed successfully", zap.String("ip", c.ClientIP()))

	utils.SetRefreshTokenCookie(c, pair.RawRefreshToken, h.jwtUtil.RefreshTokenTTL(), h.cookieCfg)

	respond(c, http.StatusOK, "Tokens refreshed", gin.H{
		"accessToken": pair.AccessToken,
		"expiresIn":   pair.ExpiresIn,
		"tokenType":   "Bearer",
	})
}

// Logout godoc
// @Summary      Logout current session
// @Description  Revokes the current refresh token and clears the cookie.
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dtos.StructuredResponse
// @Failure      401  {object}  dtos.StructuredResponse
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	rawToken, _ := c.Cookie(utils.RefreshTokenCookieName)

	if rawToken != "" {
		if err := h.authSvc.Logout(c.Request.Context(), rawToken); err != nil {
			h.logger.Error("logout error", zap.String("error", err.Error()))

		} else {
			h.logger.Info("user logged out successfully")
		}
	}

	utils.ClearRefreshTokenCookie(c, h.cookieCfg)
	respond(c, http.StatusOK, "Logged out successfully", nil)
}

// LogoutAll godoc
// @Summary      Logout all sessions
// @Description  Revokes all refresh tokens for the authenticated user (all devices).
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dtos.StructuredResponse
// @Failure      401  {object}  dtos.StructuredResponse
// @Router       /auth/logout-all [post]
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		h.logger.Warn("logout-all unauthorized attempt")
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}
	h.logger.Info("logout-all attempt", zap.String("user_id", userID.String()))

	if err := h.authSvc.LogoutAll(c.Request.Context(), userID); err != nil {
		h.logger.Error("logout-all failed", zap.Error(err))
		respondError(c, http.StatusInternalServerError, "failed to revoke sessions")
		return
	}

	h.logger.Info("all sessions revoked", zap.String("user_id", userID.String()))

	utils.ClearRefreshTokenCookie(c, h.cookieCfg)
	respond(c, http.StatusOK, "All sessions terminated", nil)
}

// Me godoc
// @Summary      Get current user
// @Description  Returns the identity of the currently authenticated user from the token claims.
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dtos.StructuredResponse
// @Failure      401  {object}  dtos.StructuredResponse
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		h.logger.Warn("unauthorized /me access attempt")
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}
	role, _ := middleware.GetUserRole(c)

	h.logger.Info("me endpoint accessed",
		zap.String("user_id", userID.String()),
		zap.String("role", role),
	)

	respond(c, http.StatusOK, "ok", gin.H{
		"id":   userID,
		"role": role,
	})
}
