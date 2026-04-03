package services

import (
	"context"
	"errors"
	"fmt"
	"kycvault/internal/dtos"
	"kycvault/internal/models"
	"kycvault/internal/repository"
	"kycvault/internal/utils"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrUserAlreadyExists   = errors.New("an account with this email already exists")
	ErrPasswordMismatch    = errors.New("passwords do not match")
	ErrRefreshTokenInvalid = errors.New("refresh token is invalid")
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
	ErrRefreshTokenReuse   = errors.New("refresh token reuse detected — all sessions have been terminated")
	ErrInternal            = errors.New("an internal error occurred")
)

// TokenPair is the value object returned after a successful auth operation.
type TokenPair struct {
	AccessToken     string
	RawRefreshToken string // Send to client; never persist this value
	ExpiresIn       int64  // Access token TTL in seconds (for the client)
}

type AuthService interface {
	Register(ctx context.Context, dto dtos.RegisterUserDto) (*models.User, error)
	Login(ctx context.Context, dto dtos.LoginUserDto, ipAddress, userAgent string) (*TokenPair, error)
	RefreshTokens(ctx context.Context, rawRefreshToken, ipAddress, userAgent string) (*TokenPair, error)
	Logout(ctx context.Context, rawRefreshToken string) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error
}

type authService struct {
	repo    repository.AuthRepository
	jwtUtil *utils.JWTUtil
	logger  *zap.Logger
}

func NewAuthService(
	repo repository.AuthRepository,
	jwtUtil *utils.JWTUtil,
	logger *zap.Logger,
) AuthService {
	return &authService{
		repo:    repo,
		jwtUtil: jwtUtil,
		logger:  logger,
	}
}

func (s *authService) Register(ctx context.Context, dto dtos.RegisterUserDto) (*models.User, error) {
	if dto.Password != dto.ConfirmPassword {
		return nil, ErrPasswordMismatch
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("service: hash password: %w", err)
	}

	user := &models.User{
		Email:        dto.Email,
		FirstName:    dto.FirstName,
		LastName:     dto.LastName,
		PasswordHash: string(hash),
		Role:         "user",
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}
		s.logger.Warn("Registration attempt with existing email",
			zap.String("email", dto.Email),
		)
		return nil, ErrInternal
	}

	s.logger.Info("User registered successfully", zap.String("email", user.Email))
	return user, nil
}

func (s *authService) Login(ctx context.Context, dto dtos.LoginUserDto, ipAddress, userAgent string) (*TokenPair, error) {
	user, err := s.repo.GetUserByEmail(ctx, dto.Email)
	if err != nil {

		_ = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(dto.Password))
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		s.logger.Error("failed to fetch user for login", zap.String("email", user.Email))
		return nil, ErrInternal
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(dto.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokenPair(ctx, user, ipAddress, userAgent)
}

func (s *authService) RefreshTokens(ctx context.Context, rawRefreshToken, ipAddress, userAgent string) (*TokenPair, error) {
	hash := utils.HashToken(rawRefreshToken)

	token, err := s.repo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRefreshTokenRevoked):
			// Reuse of a revoked token is a strong signal of token theft.
			// Revoke the entire family as a precaution.
			s.logger.Warn("revoked refresh token reuse detected, revoking all user sessions",
				zap.String("hash", hash))
			if token != nil {
				_ = s.repo.RevokeAllUserRefreshTokens(ctx, token.UserID)
			}
			return nil, ErrRefreshTokenReuse
		case errors.Is(err, repository.ErrRefreshTokenExpired):
			return nil, ErrRefreshTokenExpired
		case errors.Is(err, repository.ErrRefreshTokenNotFound):
			return nil, ErrRefreshTokenInvalid
		default:
			s.logger.Error("refresh token lookup failed", zap.String("error", err.Error()))
			return nil, ErrInternal
		}
	}

	// Rotate: revoke the used token immediately before issuing a new one.
	if err := s.repo.RevokeRefreshToken(ctx, token.ID); err != nil {
		s.logger.Error("failed to revoke used refresh token", zap.String("error", err.Error()))
		return nil, ErrInternal
	}

	return s.issueTokenPair(ctx, &token.User, ipAddress, userAgent)
}

func (s *authService) Logout(ctx context.Context, rawRefreshToken string) error {
	hash := utils.HashToken(rawRefreshToken)

	token, err := s.repo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		// Token not found or already revoked treat as a successful logout.
		return nil
	}

	if err := s.repo.RevokeRefreshToken(ctx, token.ID); err != nil {
		s.logger.Error("failed to revoke refresh token on logout", zap.String("error", err.Error()))
		return ErrInternal
	}

	s.logger.Info("user logged out", zap.Any("user_id", token.UserID))
	return nil
}

func (s *authService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.RevokeAllUserRefreshTokens(ctx, userID); err != nil {
		s.logger.Error("failed to revoke all user tokens",
			zap.Any("user_id", userID),
			zap.String("error", err.Error()))
		return ErrInternal
	}
	s.logger.Info("all sessions revoked", zap.Any("user_id", userID))
	return nil
}

// this generates a fresh access and refresh token and persists the refresh token hash
func (s *authService) issueTokenPair(ctx context.Context, user *models.User, ipAddress, userAgent string) (*TokenPair, error) {
	accessToken, err := s.jwtUtil.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("service: generate access token: %w", err)
	}

	rawRefresh, refreshHash, err := utils.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("service: generate refresh token: %w", err)
	}

	ttl := s.jwtUtil.RefreshTokenTTL()
	rtRecord := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: time.Now().UTC().Add(ttl),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	if err := s.repo.CreateRefreshToken(ctx, rtRecord); err != nil {
		return nil, fmt.Errorf("service: persist refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:     accessToken,
		RawRefreshToken: rawRefresh,
		ExpiresIn:       int64(s.jwtUtil.AccessTokenTTL().Seconds()),
	}, nil
}
