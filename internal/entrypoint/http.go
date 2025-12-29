package entrypoint

import (
	"task-pool/internal/entrypoint/handler"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/swagger/v2"
)

type HandlerOptions struct {
	TaskHandler *handler.TaskHandler
}

func RegisterHttpHandlers(app *fiber.App, options HandlerOptions) {
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Swagger
	app.Get("/swagger.json", func(c fiber.Ctx) error {
		return c.SendFile("docs/swagger.json")
	})

	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL: "/swagger.json",
	}))

	apiV1 := app.Group("/api/v1")
	taskGroup := apiV1.Group("/tasks")
	{
		taskGroup.Post("", options.TaskHandler.CreateTask)
		taskGroup.Get("", options.TaskHandler.GetAllTasks)
		taskGroup.Get("/:id", options.TaskHandler.GetTaskByID)
	}
}
