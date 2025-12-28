package postgres

import (
	"context"
	"errors"
	"fmt"
	"task-pool/internal/domain/entity"
	"task-pool/internal/domain/repository"

	"gorm.io/gorm"
)

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) repository.TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) model(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Model(&entity.Task{})
}

func (r *taskRepository) Create(ctx context.Context, task *entity.Task) error {
	err := r.model(ctx).Create(task).Error
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

func (r *taskRepository) FindByID(ctx context.Context, id uint64) (*entity.Task, error) {
	var task entity.Task

	err := r.model(ctx).Where("id = ?", id).First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrTaskNotFound
		}

		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

func (r *taskRepository) FindAll(ctx context.Context) ([]*entity.Task, error) {
	var tasks []*entity.Task

	err := r.model(ctx).Find(&tasks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	return tasks, nil
}

func (r *taskRepository) Update(ctx context.Context, task *entity.Task) error {
	err := r.model(ctx).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"title":       task.Title,
		"description": task.Description,
		"status":      task.Status,
	}).Error
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}
