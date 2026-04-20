package main

import (
	"fintech/internal/database"
	"fintech/internal/routes"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {

	// 1. Connect to db

	db := database.ConnectDB()

	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     getAllowedOrigins(),
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, app-id, app-secret",
		ExposeHeaders:    "Content-Length",
		AllowCredentials: true,
	}))

	routes.Setup(app, db)

	log.Fatal(app.Listen(":8000"))
	// log.Fatal(app.Listen("localhost:8000"))

}

func getAllowedOrigins() string {
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		return origins
	}

	return "http://localhost:3000,http://127.0.0.1:3000,http://localhost:5173,http://127.0.0.1:5173,http://localhost:8000,http://127.0.0.1:8000"
}
