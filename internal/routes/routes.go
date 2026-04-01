package routes

import (
	"fintech/internal/handlers"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	app.Get("/", handlers.GetHello)
}
