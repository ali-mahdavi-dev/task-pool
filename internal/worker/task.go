package worker

import (
	"context"
	"task-pool/config"
	"task-pool/internal/domain/entity"
	"task-pool/internal/domain/repository"
)

type taskWorker[T any] struct {
	config         config.Config
	taskChannel    chan *entity.Task
	taskRepository repository.TaskRepository
}

func NewTaskWorker(
	taskRepository repository.TaskRepository,
	config config.Config,
	taskChannel chan *entity.Task,
) Worker[*entity.Task] {
	return &taskWorker[*entity.Task]{
		config:         config,
		taskChannel:    taskChannel,
		taskRepository: taskRepository,
	}
}

func (w *taskWorker[T]) Run(ctx context.Context) {
	for i := 0; i < w.config.TaskWorker.Workers; i++ {
		go w.wroker(ctx)
	}
}

func (w *taskWorker[T]) wroker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-w.taskChannel:
			w.handle(ctx, task)
		}
	}
}

func (w *taskWorker[T]) handle(ctx context.Context, command *entity.Task) {
	
}
