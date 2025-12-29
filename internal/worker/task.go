package worker

import (
	"context"
	"math/rand"
	"sync"
	"task-pool/internal/domain/entity"
	"task-pool/internal/domain/repository"
	"task-pool/pkg/logger"
	"time"
)

type taskWorker[T any] struct {
	ctx            context.Context
	workersCount            uint64
	taskChannel    chan *entity.Task
	taskRepository repository.TaskRepository
	wg             sync.WaitGroup
}

func NewTaskWorker(
	taskRepository repository.TaskRepository,
	workersCount uint64,
	taskChannel chan *entity.Task,
) Worker[*entity.Task] {
	return &taskWorker[*entity.Task]{
		workersCount:            workersCount,
		taskChannel:    taskChannel,
		taskRepository: taskRepository,
		wg:             sync.WaitGroup{},
	}
}

func (w *taskWorker[T]) Run(ctx context.Context) {
	w.ctx = ctx
	for i := 0; i < int(w.workersCount); i++ {
		w.wg.Add(1)
		go w.wroker(ctx)
	}
}

func (w *taskWorker[T]) Shutdown() {}

func (w *taskWorker[T]) wroker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.wg.Done()
			return
		case task := <-w.taskChannel:
			w.handle(ctx, task)
		}
	}
}

func (w *taskWorker[T]) handle(ctx context.Context, command *entity.Task) {
	logger.Info("Starting task processing").
		WithUint64("task_id", command.ID).
		WithString("task_title", command.Title).
		Log()

	num := rand.Intn(5) + 1
	duration := time.Duration(num) * time.Second

	time.Sleep(duration)

	command.Complete()
	err := w.taskRepository.Update(ctx, command)
	if err != nil {
		logger.Error("Error creating task").WithError(err).Log()
		return
	}

	logger.Info("Task completed successfully").WithUint64("task_id", command.ID).Log()
}
