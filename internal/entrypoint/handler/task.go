package handler

import (
	"strconv"
	"task-pool/internal/service/contracts"
	"task-pool/pkg/apperror"

	"github.com/gofiber/fiber/v3"
)

type TaskHandler struct {
	taskService contracts.TaskService
}

func NewTaskHandler(taskService contracts.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) CreateTask(c fiber.Ctx) error {
	var command contracts.CreateTask
	if err := c.Bind().Body(&command); err != nil {
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

func (h *TaskHandler) GetTaskByID(c fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return apperror.HandleError(c, err)
	}

	task, err := h.taskService.GetByID(c.Context(), id)
	if err != nil {
		return apperror.HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(task)
}

func (h *TaskHandler) GetAllTasks(c fiber.Ctx) error {
	tasks, err := h.taskService.GetAll(c.Context())
	if err != nil {
		return apperror.HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(tasks)
}
