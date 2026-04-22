package bills

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"gorm.io/gorm"
)

type AuthService struct {
	DB *gorm.DB
}

func GenerateSecureReference() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", errors.New("unable to generate secure random reference")
	}

	return hex.EncodeToString(randomBytes), nil
}
