package wallet

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

func GetBalance(c *fiber.Ctx) error {
	// fiber JWT middleware automatically stores the user info in 'c.Locals'

	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	userID := claims["user_id"].(string)

	return c.JSON(fiber.Map{
		"message": "Welcome!",
		"user_id": userID,
	})
}
