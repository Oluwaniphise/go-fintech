package main

import (
	"fintech/internal/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// Register routes
	routes.Setup(app)

	// start server
	app.Listen(":3000")
}
