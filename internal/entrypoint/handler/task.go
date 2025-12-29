package handler

import (
	"strconv"
	_ "task-pool/internal/domain/entity" // for swagger docs
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

// CreateTask creates a new task
//
//	@Summary		Create a new task
//	@Description	Create a new task with title and description
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			task	body		contracts.CreateTask	true	"Task creation request"
//	@Success		201		{object}	map[string]string		"Task created successfully"
//	@Failure		400		{object}	map[string]string		"Bad request - invalid input"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Router			/api/v1/tasks [post]
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

// GetTaskByID retrieves a task by its ID
//
//	@Summary		Get task by ID
//	@Description	Get a specific task by its unique identifier
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			id	path		uint64				true	"Task ID"
//	@Success		200	{object}	entity.Task			"Task details"
//	@Failure		400	{object}	map[string]string	"Bad request - invalid ID format"
//	@Failure		404	{object}	map[string]string	"Task not found"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Router			/api/v1/tasks/{id} [get]
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

// GetAllTasks retrieves all tasks
//
//	@Summary		Get all tasks
//	@Description	Get a list of all tasks in the system
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		entity.Task			"List of tasks"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Router			/api/v1/tasks [get]
func (h *TaskHandler) GetAllTasks(c fiber.Ctx) error {
	tasks, err := h.taskService.GetAll(c.Context())
	if err != nil {
		return apperror.HandleError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(tasks)
}
