package auth

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func shouldLogAuthTokens() bool {
	return strings.EqualFold(os.Getenv("LOG_AUTH_TOKENS"), "true")
}

func DebugTokenLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if shouldLogAuthTokens() {
			accessToken := c.Cookies(accessCookieName)
			refreshToken := c.Cookies(refreshCookieName)

			if accessToken != "" || refreshToken != "" {
				log.Printf(
					"incoming auth cookies method=%s path=%s access_token=%q refresh_token=%q",
					c.Method(),
					c.Path(),
					accessToken,
					refreshToken,
				)
			}
		}

		return c.Next()
	}
}
