package worker

import (
	"context"
	"errors"
	"task-pool/config"
	"task-pool/internal/domain/entity"
	testmock "task-pool/test/mock"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaskWorker_Handle(t *testing.T) {
	ctx := context.Background()

	t.Run("successful task handling", func(t *testing.T) {
		mockRepo := testmock.NewTaskRepository()
		taskChannel := make(chan *entity.Task, 10)
		cfg := config.Config{
			TaskWorker: config.TaskWorker{
				Workers: 1,
			},
		}

		worker := NewTaskWorker(mockRepo, uint64(cfg.TaskWorker.Workers), taskChannel).(*taskWorker[*entity.Task])

		task := &entity.Task{
			ID:          1,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      entity.TaskStatusPending,
		}

		mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(updatedTask *entity.Task) bool {
			return updatedTask.ID == task.ID &&
				updatedTask.Status == entity.TaskStatusCompleted
		})).Return(nil)

		worker.handle(ctx, task)

		assert.Equal(t, entity.TaskStatusCompleted, task.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error on update", func(t *testing.T) {
		mockRepo := testmock.NewTaskRepository()
		taskChannel := make(chan *entity.Task, 10)
		cfg := config.Config{
			TaskWorker: config.TaskWorker{
				Workers: 1,
			},
		}

		worker := NewTaskWorker(mockRepo, uint64(cfg.TaskWorker.Workers), taskChannel).(*taskWorker[*entity.Task])

		task := &entity.Task{
			ID:          1,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      entity.TaskStatusPending,
		}

		expectedError := errors.New("database connection failed")
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(expectedError)

		worker.handle(ctx, task)

		// Task should still be completed even if update fails
		assert.Equal(t, entity.TaskStatusCompleted, task.Status)
		mockRepo.AssertExpectations(t)
	})
}

func TestTaskWorker_Run(t *testing.T) {
	ctx := context.Background()

	t.Run("worker starts with correct number of goroutines", func(t *testing.T) {
		mockRepo := testmock.NewTaskRepository()
		taskChannel := make(chan *entity.Task, 10)
		cfg := config.Config{
			TaskWorker: config.TaskWorker{
				Workers: 1,
			},
		}

		worker := NewTaskWorker(mockRepo, uint64(cfg.TaskWorker.Workers), taskChannel).(*taskWorker[*entity.Task])

		worker.Run(ctx)

		// Give workers time to start
		time.Sleep(100 * time.Millisecond)

		// Verify workers are running by checking if they can process tasks
		task := &entity.Task{
			ID:          1,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      entity.TaskStatusPending,
		}

		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

		taskChannel <- task

		// Wait for task to be processed (handle sleeps 1-5 seconds)
		time.Sleep(6 * time.Second)

		assert.Equal(t, entity.TaskStatusCompleted, task.Status)
		mockRepo.AssertExpectations(t)

		// Cleanup
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		worker.wroker(cancelCtx)
		worker.wg.Wait()
	})
}

func TestTaskWorker_wroker(t *testing.T) {
	t.Run("worker processes tasks from channel", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := testmock.NewTaskRepository()
		taskChannel := make(chan *entity.Task, 10)
		cfg := config.Config{
			TaskWorker: config.TaskWorker{
				Workers: 1,
			},
		}

		worker := NewTaskWorker(mockRepo, uint64(cfg.TaskWorker.Workers), taskChannel).(*taskWorker[*entity.Task])

		task := &entity.Task{
			ID:          1,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      entity.TaskStatusPending,
		}

		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

		// Start worker in goroutine
		worker.wg.Add(1)
		go worker.wroker(ctx)

		// Send task to channel
		taskChannel <- task

		// Wait for task to be processed (handle sleeps 1-5 seconds)
		time.Sleep(6 * time.Second)

		assert.Equal(t, entity.TaskStatusCompleted, task.Status)
		mockRepo.AssertExpectations(t)

		// Cleanup: cancel context to stop worker
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		worker.wroker(cancelCtx)
		worker.wg.Wait()
	})

	t.Run("worker stops on context cancellation", func(t *testing.T) {
		mockRepo := testmock.NewTaskRepository()
		taskChannel := make(chan *entity.Task, 10)
		cfg := config.Config{
			TaskWorker: config.TaskWorker{
				Workers: 1,
			},
		}

		worker := NewTaskWorker(mockRepo, uint64(cfg.TaskWorker.Workers), taskChannel).(*taskWorker[*entity.Task])

		ctx, cancel := context.WithCancel(context.Background())

		// Start worker
		worker.wg.Add(1)
		go worker.wroker(ctx)

		// Give worker time to start
		time.Sleep(100 * time.Millisecond)

		// Cancel context
		cancel()

		// Wait for worker to finish
		done := make(chan bool)
		go func() {
			worker.wg.Wait()
			done <- true
		}()

		select {
		case <-done:
			// Worker stopped successfully
		case <-time.After(2 * time.Second):
			t.Fatal("worker did not stop on context cancellation")
		}
	})

	t.Run("worker processes multiple tasks", func(t *testing.T) {
		ctx := context.Background()
		mockRepo := testmock.NewTaskRepository()
		taskChannel := make(chan *entity.Task, 10)
		cfg := config.Config{
			TaskWorker: config.TaskWorker{
				Workers: 1,
			},
		}

		worker := NewTaskWorker(mockRepo, uint64(cfg.TaskWorker.Workers), taskChannel).(*taskWorker[*entity.Task])

		task1 := &entity.Task{
			ID:          1,
			Title:       "Task 1",
			Description: "Description 1",
			Status:      entity.TaskStatusPending,
		}

		task2 := &entity.Task{
			ID:          2,
			Title:       "Task 2",
			Description: "Description 2",
			Status:      entity.TaskStatusPending,
		}

		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Twice()

		// Start worker
		worker.wg.Add(1)
		go worker.wroker(ctx)

		// Send tasks to channel
		taskChannel <- task1
		taskChannel <- task2

		// Wait for tasks to be processed (handle sleeps 1-5 seconds per task)
		time.Sleep(12 * time.Second)

		assert.Equal(t, entity.TaskStatusCompleted, task1.Status)
		assert.Equal(t, entity.TaskStatusCompleted, task2.Status)
		mockRepo.AssertExpectations(t)

		// Cleanup
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()
		worker.wroker(cancelCtx)
		worker.wg.Wait()
	})
}

