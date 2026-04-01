package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrTokenExpired   = errors.New("token has expired")
	ErrTokenInvalid   = errors.New("token is invalid")
	ErrTokenMalformed = errors.New("token is malformed")
)

// accesstokenclaims are claims inserted in a short lived access token, sensitive data are excluded
type AccessTokenClaims struct {
	UserID uuid.UUID `json:"uuid"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// holds all config required to issue and validate tokesn
type JWTConfig struct {
	AccessSecret    string        // HS256 secret for access tokens
	AccessTokenTTL  time.Duration // 20 mins
	RefreshTokenTTL time.Duration // 7 days
	Issuer          string        // e.g. "api.yourapp.com"
}

// JWTUtil encapsulates all JWT operations. Construct once and share.
type JWTUtil struct {
	cfg JWTConfig
}

// NewJWTUtil validates the config and returns a ready-to-use JWTUtil.
func NewJWTUtil(cfg JWTConfig) (*JWTUtil, error) {
	if cfg.AccessSecret == "" {
		return nil, errors.New("jwt: AccessSecret must not be empty")
	}
	if len(cfg.AccessSecret) < 32 {
		return nil, errors.New("jwt: AccessSecret must be at least 32 characters for HS256 security")
	}
	if cfg.AccessTokenTTL <= 0 {
		return nil, errors.New("jwt: AccessTokenTTL must be positive")
	}
	if cfg.RefreshTokenTTL <= 0 {
		return nil, errors.New("jwt: RefreshTokenTTL must be positive")
	}
	return &JWTUtil{cfg: cfg}, nil
}

// GenerateAccessToken issues a signed access token for the user.
func (j *JWTUtil) GenerateAccessToken(userID uuid.UUID, role string) (string, error) {
	now := time.Now().UTC()
	claims := AccessTokenClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.cfg.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.cfg.AccessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(j.cfg.AccessSecret))
	if err != nil {
		return "", fmt.Errorf("jwt: failed to sign access token: %w", err)
	}
	return signed, nil
}

// ValidateAccessToken parses and validates the token string.
func (j *JWTUtil) ValidateAccessToken(tokenStr string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&AccessTokenClaims{},
		func(t *jwt.Token) (interface{}, error) {
			// reject any token not using HS256.
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("jwt: unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(j.cfg.AccessSecret), nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithIssuedAt(),
	)

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, ErrTokenExpired
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, ErrTokenMalformed
		default:
			return nil, ErrTokenInvalid
		}
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

// GenerateRefrshtoken returns a random token string and the sha hash, we store only the hash and send the raw to client with httponly cookie
func GenerateRefreshToken() (rawToken string, tokenHash string, err error) {
	b := make([]byte, 64) // 512 bits of entropy
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("jwt: failed to generate refresh token: %w", err)
	}
	rawToken = hex.EncodeToString(b)
	tokenHash = HashToken(rawToken)
	return rawToken, tokenHash, nil
}

// this produces sha-256 hex of the token string and we use it both when storing and looking for refresh token
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// thsi exposes the refersh token lifetime
func (j *JWTUtil) RefreshTokenTTL() time.Duration {
	return j.cfg.RefreshTokenTTL
}

// this exposes the access token lifetime
func (j *JWTUtil) AccessTokenTTL() time.Duration {
	return j.cfg.AccessTokenTTL
}
