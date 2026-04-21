// internal/auth/tokens.go
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"os"
	"time"

	"fintech/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

const (
	refreshCookieName = "refresh_token"

	accessTokenTTL  = 500 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

func (s *AuthService) SignAccessToken(user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID.String(),
		"type":    "access",
		"exp":     time.Now().Add(accessTokenTTL).Unix(),
		"iat":     time.Now().Unix(),
	})

	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func generateRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
