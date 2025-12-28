package entrypoint

import (
	"task-pool/internal/entrypoint/handler"

	"github.com/gofiber/fiber/v2"
)

type HandlerOptions struct {
	TaskHandler *handler.TaskHandler
}

func RegisterHttpHandlers(app *fiber.App, options HandlerOptions) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Post("/tasks", options.TaskHandler.CreateTask)
}
