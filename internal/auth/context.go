package auth

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidTokenClaims = errors.New("invalid token claims")
	ErrUserIDMissing      = errors.New("user_id missing in token")
)

func GetUserIDFromContext(c *fiber.Ctx) (string, error) {
	userToken, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return "", ErrUnauthorized
	}

	claims, ok := userToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidTokenClaims
	}

	userID, ok := claims["user_id"]
	if !ok {
		return "", ErrUserIDMissing
	}

	return fmt.Sprint(userID), nil
}
