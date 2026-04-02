package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	RefreshTokenCookieName = "refresh_token"
	cookiePath             = "/api/v1/auth" // Scoped: only sent on auth routes
)

// CookieConfig controls the security attributes of the refresh token cookie.
type CookieConfig struct {
	Domain   string
	Secure   bool // Must be true in production (requires HTTPS)
	SameSite http.SameSite
}

// this writes the refresh token as httponly, secure, samesite = stricct, sent to cookiepath
func SetRefreshTokenCookie(c *gin.Context, rawToken string, ttl time.Duration, cfg CookieConfig) {
	sameSite := cfg.SameSite
	c.SetSameSite(sameSite)
	c.SetCookie(
		RefreshTokenCookieName,
		rawToken,
		int(ttl.Seconds()),
		cookiePath,
		cfg.Domain,
		cfg.Secure, // Secure flag — enforce HTTPS in production
		true,       // httpOnly — JS cannot access this cookie
	)
}

// this clears refresh token cookie immediately it expires looging out the client session
func ClearRefreshTokenCookie(c *gin.Context, cfg CookieConfig) {
	sameSite := cfg.SameSite
	c.SetSameSite(sameSite)
	c.SetCookie(
		RefreshTokenCookieName,
		"",
		-1,
		cookiePath,
		cfg.Domain,
		cfg.Secure,
		true,
	)
}

func parseSameSite(s string) http.SameSite {
	switch s {
	case "Lax":
		return http.SameSiteLaxMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteStrictMode
	}
}
