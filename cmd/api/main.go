package main

import (
	"fintech/internal/database"
	"fintech/internal/routes"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {

	// 1. Connect to db

	db := database.ConnectDB()

	app := fiber.New()

	app.Use(logger.New())

	routes.Setup(app, db)

	log.Fatal(app.Listen(":8000"))
}
