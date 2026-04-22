package routes

import (
	"fintech/internal/auth"
	"fintech/internal/bills"
	"fintech/internal/bills/airtime"
	"fintech/internal/bills/catalog"
	"fintech/internal/bills/electricity"
	"fintech/internal/transactions"
	"fintech/internal/wallet"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Setup(app *fiber.App, db *gorm.DB) {
	//1. initialize services

	deps := bills.Dependencies{
		DB:         db,
		HTTPClient: &http.Client{},
	}

	airtimeService := airtime.NewService(deps)
	// dataService := data.NewService(deps)
	catalogService := catalog.NewService(deps)
	electricityService := electricity.NewService(deps)

	authService := &auth.AuthService{DB: db}
	walletService := &wallet.WalletService{DB: db}
	transactionService := &transactions.TransactionService{DB: db}

	//2. Create a group for API versioning
	api := app.Group("/api/v1")

	// 3. auth routes
	authGroup := api.Group("/auth")
	authGroup.Post("/register", authService.HandleRegister)
	authGroup.Post("/verify-email", authService.HandleVerifyEmail)
	authGroup.Post("/resend-verification-email", authService.HandleResendVerificationEmail)
	authGroup.Post("/login", authService.HandleLogin)
	authGroup.Post("/logout", authService.HandleLogout)
	authGroup.Post("/refresh", authService.HandleRefresh)
	authGroup.Post("/login/verify-otp", authService.HandleVerifyLoginOTP)
	authGroup.Post("/login/resend-otp", authService.HandleResendLoginOTP)
	authGroup.Get("/me", ProtectedRoute(), authService.HandleLoggedInUserDetails)

	walletGroup := api.Group("/wallet", ProtectedRoute())
	billsGroup := api.Group("/bills", ProtectedRoute())
	transactionsGroup := api.Group("/transactions", ProtectedRoute())

	walletGroup.Get("/balance", walletService.HandleGetBalance)
	walletGroup.Post("/credit", walletService.HandleCreditWallet)
	billsGroup.Post("/airtime", airtimeService.HandleAirtimePurchase)
	billsGroup.Post("/electricity/validate", electricityService.HandleValidateElectricityPurchase)
	billsGroup.Post("/electricity/pay", electricityService.HandleElectricityPurchase)
	billsGroup.Get("/services/:serviceCode/items", catalogService.HandleGetBondProducts)
	billsGroup.Get("/services/items/:id", catalogService.HandleGetBondProductItems)

	transactionsGroup.Get("/me", transactionService.HandleGetUserTransactions)
	transactionsGroup.Get("/me/stats", transactionService.HandleGetUserTransactionStats)
}
