package routes

import (
	"fintech/internal/auth"
	"fintech/internal/bills"
	"fintech/internal/wallet"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	//1. initialize services

	authService := &auth.AuthService{DB: db}
	walletService := &wallet.WalletService{DB: db}
	billsService := &bills.BillService{DB: db}

	//2. Create a group for API versioning
	api := app.Group("/api/v1")

	// 3. auth routes
	authGroup := api.Group("/auth")
	authGroup.Post("/register", authService.HandleRegister)
	authGroup.Post("/login", authService.HandleLogin)

	walletGroup := api.Group("/wallet", ProtectedRoute())
	billsGroup := api.Group("/bills", ProtectedRoute())

	walletGroup.Get("/balance", walletService.GetBalance)
	walletGroup.Post("/credit", walletService.HandleCreditWallet)
	billsGroup.Post("/airtime", billsService.HandleAirtimePurchase)
}
