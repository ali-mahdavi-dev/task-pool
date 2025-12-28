package contracts

import (
	"context"
	"task-pool/internal/domain/entity"
)

type TaskService interface {
	// Create creates a new task
	Create(ctx context.Context, task *CreateTask) error

	// GetByID returns a task by its ID
	GetByID(ctx context.Context, id uint64) (*entity.Task, error)

	// GetAll returns all tasks
	GetAll(ctx context.Context) ([]*entity.Task, error)
}

type CreateTask struct {
	Title       string `json:"title" validate:"required,min=3,max=255"`
	Description string `json:"description" validate:"required,min=3,max=255"`
}
