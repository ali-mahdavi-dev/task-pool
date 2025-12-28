package service

import (
	"context"
	"errors"
	"fmt"
	"task-pool/internal/domain/entity"
	"task-pool/internal/domain/repository"
	"task-pool/internal/service/contracts"
	"task-pool/pkg/apperror"
)

type taskService struct {
	taskChannel    chan *entity.Task
	taskRepository repository.TaskRepository
}

func NewTaskService(taskRepository repository.TaskRepository) contracts.TaskService {
	return &taskService{taskRepository: taskRepository}
}

func (s *taskService) Create(ctx context.Context, command *contracts.CreateTask) error {
	task := entity.NewTask(command.Title, command.Description, entity.TaskStatusPending)
	err := s.taskRepository.Create(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	s.taskChannel <- task

	return nil
}

func (s *taskService) GetByID(ctx context.Context, id string) (*entity.Task, error) {
	task, err := s.taskRepository.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrTaskNotFound) {
			return nil, apperror.NotFound("task not found")
		}

		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}
