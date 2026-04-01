package routes

import (
	"fintech/internal/auth"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	// initialize services

	authService := &auth.AuthService{DB: db}

	// Create a group for API versioning
	api := app.Group("/api/v1")

	// 3 auth routes
	authGroup := api.Group("/auth")
	authGroup.Post("/register", authService.HandleRegister)
	authGroup.Post("/login", authService.HandleLogin)
}
