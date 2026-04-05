package repository

import (
	"context"
	"errors"
	"fmt"
	"kycvault/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrUserAlreadyExists    = errors.New("user with that email already exists")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenRevoked  = errors.New("refresh token has been revoked")
	ErrRefreshTokenExpired  = errors.New("refresh token has expired")
)

type AuthRepository interface {
	// User
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)

	// Refresh token
	CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error
	GetRefreshTokenByHash(ctx context.Context, hash string) (*models.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredTokens(ctx context.Context) (int64, error)
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

// FOR USERS
func (r *authRepository) CreateUser(ctx context.Context, user *models.User) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		if IsDuplicateKeyError(result.Error) {
			return ErrUserAlreadyExists
		}
		return fmt.Errorf("repository: create user: %w", result.Error)
	}
	return nil
}

func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository: get user by email: %w", result.Error)
	}
	return &user, nil
}

func (r *authRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("repository: get user by id: %w", result.Error)
	}
	return &user, nil
}

// FOR REFRESH
func (r *authRepository) CreateRefreshToken(ctx context.Context, token *models.RefreshToken) error {
	result := r.db.WithContext(ctx).Create(token)
	if result.Error != nil {
		return fmt.Errorf("repository: create refresh token: %w", result.Error)
	}
	return nil
}

// this fetches and validates a token
func (r *authRepository) GetRefreshTokenByHash(ctx context.Context, hash string) (*models.RefreshToken, error) {
	var token models.RefreshToken
	result := r.db.WithContext(ctx).
		Preload("User").
		Where("token_hash = ?", hash).
		First(&token)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("repository: get refresh token: %w", result.Error)
	}

	if token.Revoked {
		return nil, ErrRefreshTokenRevoked
	}
	if time.Now().UTC().After(token.ExpiresAt) {
		return nil, ErrRefreshTokenExpired
	}

	return &token, nil
}

// RevokeRefreshToken marks a single token as revoked without deleting it,
func (r *authRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("id = ? AND revoked = false", id).
		Update("revoked", true)

	if result.Error != nil {
		return fmt.Errorf("repository: revoke refresh token: %w", result.Error)
	}
	return nil
}

// this invalidates every active session for a user called on password chane, logout everybody
func (r *authRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Model(&models.RefreshToken{}).
		Where("user_id = ? AND revoked = false", userID).
		Update("revoked", true)

	if result.Error != nil {
		return fmt.Errorf("repository: revoke all user refresh tokens: %w", result.Error)
	}
	return nil
}

// This deletes tokens past their expiry time that are already revoked
func (r *authRepository) DeleteExpiredTokens(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at < ? AND revoked = true", time.Now().UTC()).
		Delete(&models.RefreshToken{})

	if result.Error != nil {
		return 0, fmt.Errorf("repository: delete expired tokens: %w", result.Error)
	}
	// Returns the number of rows deleted.
	return result.RowsAffected, nil
}
