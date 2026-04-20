// internal/auth/cookies.go
package auth

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
)

func setAuthCookies(c *fiber.Ctx, accessToken, refreshToken string) {
	secure := os.Getenv("APP_ENV") == "production"

	c.Cookie(&fiber.Cookie{
		Name:     accessCookieName,
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/api/v1",
		MaxAge:   int(accessTokenTTL.Seconds()),
	})

	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/api/v1",
		MaxAge:   int(refreshTokenTTL.Seconds()),
	})

	if shouldLogAuthTokens() {
		log.Printf("issued auth cookies path=%s access_token=%q refresh_token=%q", c.Path(), accessToken, refreshToken)
	}
}

func clearAuthCookies(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     accessCookieName,
		Value:    "",
		HTTPOnly: true,
		Path:     "/",
		MaxAge:   -1,
	})

	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		HTTPOnly: true,
		Path:     "/api/v1/",
		MaxAge:   -1,
	})
}
