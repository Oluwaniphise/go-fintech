package routes

import (
	"os"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
)

// func ProtectedRoute() fiber.Handler {
// 	return jwtware.New(jwtware.Config{
// 		SigningKey: jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
// 		ErrorHandler: func(c *fiber.Ctx, err error) error {
// 			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized: Please login to continue"})
// 		},
// 	})
// }

func ProtectedRoute() fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey:  jwtware.SigningKey{Key: []byte(os.Getenv("JWT_SECRET"))},
		TokenLookup: "cookie:access_token",
		SuccessHandler: func(c *fiber.Ctx) error {
			return c.Next()
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Please login to continue",
			})
		},
	})
}
