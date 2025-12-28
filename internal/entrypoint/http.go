package entrypoint

import (
	"task-pool/internal/entrypoint/handler"

	"github.com/gofiber/fiber/v2"
)

type HandlerOptions struct {
	TaskHandler *handler.TaskHandler
}

func RegisterHttpHandlers(app *fiber.App, options HandlerOptions) {
	app.Post("/tasks", options.TaskHandler.CreateTask)
}
