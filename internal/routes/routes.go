package routes

import (
	"fintech/internal/auth"
	"fintech/internal/wallet"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	//1. initialize services

	authService := &auth.AuthService{DB: db}

	//2. Create a group for API versioning
	api := app.Group("/api/v1")

	// 3. auth routes
	authGroup := api.Group("/auth")
	authGroup.Post("/register", authService.HandleRegister)
	authGroup.Post("/login", authService.HandleLogin)

	walletGroup := api.Group("/wallet", ProtectedRoute())

	walletGroup.Get("/balance", wallet.GetBalance)
}
