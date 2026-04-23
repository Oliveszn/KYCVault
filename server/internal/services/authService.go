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
	ErrInvalidCredentials  = errors.New("Invalid email or password")
	ErrUserAlreadyExists   = errors.New("An account with this email already exists")
	ErrPasswordMismatch    = errors.New("Passwords do not match")
	ErrRefreshTokenInvalid = errors.New("Refresh token is invalid")
	ErrRefreshTokenExpired = errors.New("Refresh token has expired")
	ErrRefreshTokenReuse   = errors.New("Refresh token reuse detected — all sessions have been terminated")
	ErrInternal            = errors.New("An internal error occurred")
)

// TokenPair is the value object returned after a successful auth operation.
type TokenPair struct {
	AccessToken     string
	RawRefreshToken string // Send to client; never persist this value
	ExpiresIn       int64  // Access token TTL in seconds (for the client)
}

type AuthService interface {
	Register(ctx context.Context, dto dtos.RegisterUserDto, ipAddress, userAgent string) (*TokenPair, error)
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

func (s *authService) Register(ctx context.Context, dto dtos.RegisterUserDto, ipAddress, userAgent string) (*TokenPair, error) {
	s.logger.Info("register service called",
		zap.String("email", dto.Email),
		zap.String("ip", ipAddress),
	)
	if dto.Password != dto.ConfirmPassword {
		s.logger.Warn("password mismatch on register",
			zap.String("email", dto.Email),
		)
		return nil, ErrPasswordMismatch
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password",
			zap.String("email", dto.Email),
			zap.Error(err),
		)
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
			s.logger.Warn("user already exists",
				zap.String("email", dto.Email),
			)
			return nil, ErrUserAlreadyExists
		}
		s.logger.Error("failed to create user",
			zap.String("email", dto.Email),
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	s.logger.Info("User registered successfully", zap.String("email", user.Email))
	return s.issueTokenPair(ctx, user, ipAddress, userAgent)
}

func (s *authService) Login(ctx context.Context, dto dtos.LoginUserDto, ipAddress, userAgent string) (*TokenPair, error) {
	s.logger.Info("login service called",
		zap.String("email", dto.Email),
		zap.String("ip", ipAddress),
	)

	user, err := s.repo.GetUserByEmail(ctx, dto.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.logger.Warn("login failed - user not found",
				zap.String("email", dto.Email),
			)
			return nil, ErrInvalidCredentials
		}
		s.logger.Error("failed to fetch user",
			zap.String("email", dto.Email),
			zap.Error(err),
		)
		return nil, ErrInternal
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(dto.Password)); err != nil {
		s.logger.Warn("login failed - invalid password",
			zap.String("email", dto.Email),
		)
		return nil, ErrInvalidCredentials
	}
	s.logger.Info("login successful",
		zap.String("email", dto.Email),
	)

	return s.issueTokenPair(ctx, user, ipAddress, userAgent)
}

func (s *authService) RefreshTokens(ctx context.Context, rawRefreshToken, ipAddress, userAgent string) (*TokenPair, error) {
	s.logger.Info("refresh token attempt",
		zap.String("ip", ipAddress),
	)
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
			s.logger.Warn("refresh token expired")
			return nil, ErrRefreshTokenExpired
		case errors.Is(err, repository.ErrRefreshTokenNotFound):
			s.logger.Warn("refresh token not found")
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

	s.logger.Info("refresh token rotated",
		zap.String("user_id", token.UserID.String()),
	)

	return s.issueTokenPair(ctx, &token.User, ipAddress, userAgent)
}

func (s *authService) Logout(ctx context.Context, rawRefreshToken string) error {
	hash := utils.HashToken(rawRefreshToken)

	token, err := s.repo.GetRefreshTokenByHash(ctx, hash)
	if err != nil {
		s.logger.Info("logout with invalid/expired token (safe ignore)")
		return nil
	}

	if err := s.repo.RevokeRefreshToken(ctx, token.ID); err != nil {
		s.logger.Error("failed to revoke refresh token",
			zap.String("user_id", token.UserID.String()),
			zap.Error(err),
		)
		return ErrInternal
	}

	s.logger.Info("user logged out",
		zap.String("user_id", token.UserID.String()),
	)
	return nil
}

func (s *authService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	s.logger.Info("logout all sessions requested",
		zap.String("user_id", userID.String()),
	)
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
	s.logger.Info("issuing token pair",
		zap.String("user_id", user.ID.String()),
		zap.String("ip", ipAddress),
	)
	accessToken, err := s.jwtUtil.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		s.logger.Error("failed to generate access token",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("service: generate access token: %w", err)
	}

	rawRefresh, refreshHash, err := utils.GenerateRefreshToken()
	if err != nil {
		s.logger.Error("failed to generate refresh token",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
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
		s.logger.Error("failed to persist refresh token",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("service: persist refresh token: %w", err)
	}

	s.logger.Info("token pair issued successfully",
		zap.String("user_id", user.ID.String()),
	)

	return &TokenPair{
		AccessToken:     accessToken,
		RawRefreshToken: rawRefresh,
		ExpiresIn:       int64(s.jwtUtil.AccessTokenTTL().Seconds()),
	}, nil
}
