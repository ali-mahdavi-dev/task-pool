package handler

import (
	"task-pool/internal/application/contracts"
	"task-pool/pkg/apperror"

	"github.com/gofiber/fiber/v2"
)

type TaskHandler struct {
	taskService contracts.TaskService
}

func NewTaskHandler(taskService contracts.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	var command contracts.CreateTask
	if err := c.BodyParser(&command); err != nil {
		return apperror.HandleError(c, err)
	}

	err := h.taskService.Create(c.Context(), &command)
	if err != nil {
		return apperror.HandleError(c, err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Task created successfully",
	})
}
