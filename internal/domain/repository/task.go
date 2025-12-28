package repository

import (
	"context"
	"errors"
	"task-pool/internal/domain/entity"
)

var (
	ErrTaskNotFound = errors.New("task not found")
)

type TaskRepository interface {
	Create(ctx context.Context, task *entity.Task) error
	Get(ctx context.Context, id string) (*entity.Task, error)
	Update(ctx context.Context, task *entity.Task) error
	Delete(ctx context.Context, id string) error
}
