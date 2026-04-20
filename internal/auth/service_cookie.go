package auth

import (
	"errors"
	"time"

	"fintech/internal/models"

	"gorm.io/gorm"
)

var ErrInvalidRefreshToken = errors.New("invalid refresh token")

type SessionTokens struct {
	AccessToken  string
	RefreshToken string
	User         models.User
}

func (s *AuthService) CreateSession(user models.User, userAgent, ipAddress string) (*SessionTokens, error) {
	accessToken, err := s.SignAccessToken(user)
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	session := models.RefreshSession{
		UserID:    user.ID,
		TokenHash: hashRefreshToken(refreshToken),
		ExpiresAt: time.Now().Add(refreshTokenTTL),
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}

	if err := s.DB.Create(&session).Error; err != nil {
		return nil, err
	}

	return &SessionTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *AuthService) RefreshSession(rawRefreshToken string, userAgent, ipAddress string) (*SessionTokens, error) {
	tokenHash := hashRefreshToken(rawRefreshToken)

	var session models.RefreshSession
	if err := s.DB.Where("token_hash = ?", tokenHash).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}

	if session.RevokedAt != nil || time.Now().After(session.ExpiresAt) {
		return nil, ErrInvalidRefreshToken
	}

	var user models.User
	if err := s.DB.Where("id = ?", session.UserID).First(&user).Error; err != nil {
		return nil, err
	}

	newAccessToken, err := s.SignAccessToken(user)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if err := s.DB.Model(&models.RefreshSession{}).
		Where("id = ?", session.ID).
		Updates(map[string]any{
			"token_hash":   hashRefreshToken(newRefreshToken),
			"last_used_at": &now,
			"expires_at":   time.Now().Add(refreshTokenTTL),
			"user_agent":   userAgent,
			"ip_address":   ipAddress,
		}).Error; err != nil {
		return nil, err
	}

	return &SessionTokens{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		User:         user,
	}, nil
}

func (s *AuthService) RevokeSession(rawRefreshToken string) error {
	tokenHash := hashRefreshToken(rawRefreshToken)
	now := time.Now()

	return s.DB.Model(&models.RefreshSession{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", &now).Error
}
