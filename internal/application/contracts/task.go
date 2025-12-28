package contracts

import (
	"context"
	"task-pool/internal/domain/entity"
)

type TaskService interface {
	Create(ctx context.Context, task *CreateTask) error
	Get(ctx context.Context, id string) (*entity.Task, error)
	// Update(ctx context.Context, task *entity.Task) error
	// Delete(ctx context.Context, id string) error
}

type CreateTask struct {
	Title       string `json:"title" validate:"required,min=3,max=255"`
	Description string `json:"description" validate:"required,min=3,max=255"`
}
