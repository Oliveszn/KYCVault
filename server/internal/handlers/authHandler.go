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
	if !h.bindJSON(c, &dto) {
		return
	}

	pair, err := h.authSvc.Register(
		c.Request.Context(),
		dto,
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrUserAlreadyExists: http.StatusConflict,
			services.ErrPasswordMismatch:  http.StatusBadRequest,
		})
		return
	}

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
	if !h.bindJSON(c, &dto) {
		return
	}

	pair, err := h.authSvc.Login(
		c.Request.Context(),
		dto,
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrInvalidCredentials: http.StatusUnauthorized,
		})
		return
	}

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
		respondError(c, http.StatusUnauthorized, "refresh token cookie is missing or invalid")
		return
	}

	pair, err := h.authSvc.RefreshTokens(
		c.Request.Context(),
		rawToken,
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		// On reuse detection, also clear the cookie so the client is fully signed out.
		if errors.Is(err, services.ErrRefreshTokenReuse) {
			utils.ClearRefreshTokenCookie(c, h.cookieCfg)
		}
		handleServiceError(c, h.logger, err, map[error]int{
			services.ErrRefreshTokenInvalid: http.StatusUnauthorized,
			services.ErrRefreshTokenExpired: http.StatusUnauthorized,
			services.ErrRefreshTokenReuse:   http.StatusUnauthorized,
		})
		return
	}

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
			// Still clear the cookie — don't leave the client in a broken state.
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
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}

	if err := h.authSvc.LogoutAll(c.Request.Context(), userID); err != nil {
		respondError(c, http.StatusInternalServerError, "failed to revoke sessions")
		return
	}

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
		respondError(c, http.StatusUnauthorized, "unauthenticated")
		return
	}
	role, _ := middleware.GetUserRole(c)

	respond(c, http.StatusOK, "ok", gin.H{
		"id":   userID,
		"role": role,
	})
}

// bindJSON attempts to bind and validate the request body. On failure it
// writes the error response and returns false so the caller can return early.
func (h *AuthHandler) bindJSON(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		respondError(c, http.StatusBadRequest, err.Error())
		return false
	}
	return true
}
