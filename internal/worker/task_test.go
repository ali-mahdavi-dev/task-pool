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

// testFixture contains all test dependencies
type testFixture struct {
	mockRepo    *testmock.TaskRepository
	taskChannel chan *entity.Task
	cfg         config.Config
	task        *entity.Task
	ctx         context.Context
	worker      *taskWorker[*entity.Task]
}

// setupFixture creates a simple test fixture with default values
func setupFixture() *testFixture {
	f := &testFixture{
		mockRepo:    testmock.NewTaskRepository(),
		taskChannel: make(chan *entity.Task, 10),
		cfg: config.Config{
			TaskWorker: config.TaskWorker{
				Workers: 1,
			},
		},
		task: &entity.Task{
			ID:          1,
			Title:       "Test Task",
			Description: "Test Description",
			Status:      entity.TaskStatusPending,
		},
		ctx: context.Background(),
	}
	
	// Create worker
	f.worker = NewTaskWorker(f.mockRepo, uint64(f.cfg.TaskWorker.Workers), f.taskChannel).(*taskWorker[*entity.Task])
	
	return f
}

func TestTaskWorker_Handle(t *testing.T) {
	t.Run("successful task handling", func(t *testing.T) {
		f := setupFixture()

		f.mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(updatedTask *entity.Task) bool {
			return updatedTask.ID == f.task.ID &&
				updatedTask.Status == entity.TaskStatusCompleted
		})).Return(nil)

		f.worker.handle(f.ctx, f.task)

		assert.Equal(t, entity.TaskStatusCompleted, f.task.Status)
		f.mockRepo.AssertExpectations(t)
	})

	t.Run("repository error on update", func(t *testing.T) {
		f := setupFixture()

		expectedError := errors.New("database connection failed")
		f.mockRepo.On("Update", mock.Anything, mock.Anything).Return(expectedError)

		f.worker.handle(f.ctx, f.task)

		// Task should still be completed even if update fails
		assert.Equal(t, entity.TaskStatusCompleted, f.task.Status)
		f.mockRepo.AssertExpectations(t)
	})
}

func TestTaskWorker_Run(t *testing.T) {
	t.Run("worker starts with correct number of goroutines", func(t *testing.T) {
		f := setupFixture()

		f.worker.Run(f.ctx)

		// Give workers time to start
		time.Sleep(100 * time.Millisecond)

		// Verify workers are running by checking if they can process tasks
		f.mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

		f.taskChannel <- f.task

		// Wait for task to be processed (handle sleeps 1-5 seconds)
		time.Sleep(6 * time.Second)

		assert.Equal(t, entity.TaskStatusCompleted, f.task.Status)
		f.mockRepo.AssertExpectations(t)

		// Cleanup
		cancelCtx, cancel := context.WithCancel(f.ctx)
		cancel()
		f.worker.wroker(cancelCtx)
		f.worker.wg.Wait()
	})
}

func TestTaskWorker_wroker(t *testing.T) {
	t.Run("worker processes tasks from channel", func(t *testing.T) {
		f := setupFixture()

		f.mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

		// Start worker in goroutine
		f.worker.wg.Add(1)
		go f.worker.wroker(f.ctx)

		// Send task to channel
		f.taskChannel <- f.task

		// Wait for task to be processed (handle sleeps 1-5 seconds)
		time.Sleep(6 * time.Second)

		assert.Equal(t, entity.TaskStatusCompleted, f.task.Status)
		f.mockRepo.AssertExpectations(t)

		// Cleanup: cancel context to stop worker
		cancelCtx, cancel := context.WithCancel(f.ctx)
		cancel()
		f.worker.wroker(cancelCtx)
		f.worker.wg.Wait()
	})

	t.Run("worker stops on context cancellation", func(t *testing.T) {
		f := setupFixture()

		ctx, cancel := context.WithCancel(context.Background())

		// Start worker
		f.worker.wg.Add(1)
		go f.worker.wroker(ctx)

		// Give worker time to start
		time.Sleep(100 * time.Millisecond)

		// Cancel context
		cancel()

		// Wait for worker to finish
		done := make(chan bool)
		go func() {
			f.worker.wg.Wait()
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
		f := setupFixture()

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

		f.mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil).Twice()

		// Start worker
		f.worker.wg.Add(1)
		go f.worker.wroker(f.ctx)

		// Send tasks to channel
		f.taskChannel <- task1
		f.taskChannel <- task2

		// Wait for tasks to be processed (handle sleeps 1-5 seconds per task)
		time.Sleep(12 * time.Second)

		assert.Equal(t, entity.TaskStatusCompleted, task1.Status)
		assert.Equal(t, entity.TaskStatusCompleted, task2.Status)
		f.mockRepo.AssertExpectations(t)

		// Cleanup
		cancelCtx, cancel := context.WithCancel(f.ctx)
		cancel()
		f.worker.wroker(cancelCtx)
		f.worker.wg.Wait()
	})
}

