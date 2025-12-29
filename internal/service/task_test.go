package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"task-pool/internal/domain/entity"
	"task-pool/internal/domain/repository"
	"task-pool/internal/service/contracts"
	"task-pool/pkg/apperror"
	testmock "task-pool/test/mock"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestTaskService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("successful task creation", func(t *testing.T) {
		mockRepo := testmock.NewTaskRepository()
		taskChannel := make(chan *entity.Task, 10)
		service := NewTaskService(mockRepo, taskChannel)

		createCmd := &contracts.CreateTask{
			Title:       "Test Task",
			Description: "Test Description",
		}

		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		err := service.Create(ctx, createCmd)
		require.NoError(t, err)

		select {
		case task := <-taskChannel:
			assert.Equal(t, createCmd.Title, task.Title)
			assert.Equal(t, createCmd.Description, task.Description)
			assert.Equal(t, entity.TaskStatusPending, task.Status)
		case <-time.After(1 * time.Second):
			t.Fatal("task was not sent to channel")
		}

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error on create", func(t *testing.T) {
		mockRepo := testmock.NewTaskRepository()
		taskChannel := make(chan *entity.Task, 10)
		service := NewTaskService(mockRepo, taskChannel)

		createCmd := &contracts.CreateTask{
			Title:       "Test Task",
			Description: "Test Description",
		}

		expectedError := errors.New("database connection failed")
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(expectedError)

		err := service.Create(ctx, createCmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create task")
		assert.Contains(t, err.Error(), "database connection failed")

		// Verify task was not sent to channel
		select {
		case <-taskChannel:
			t.Fatal("task should not be sent to channel when repository fails")
		default:
			// Expected: channel should be empty
		}

		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("successful task retrieval by id", func(t *testing.T) {
		mockRepo := testmock.NewTaskRepository()
		taskChannel := make(chan *entity.Task, 10)
		service := NewTaskService(mockRepo, taskChannel)

		createCmd := &contracts.CreateTask{
			Title:       "Test Task",
			Description: "Test Description",
		}

		expectedTask := &entity.Task{
			ID:          1,
			Title:       createCmd.Title,
			Description: createCmd.Description,
			Status:      entity.TaskStatusPending,
		}
		mockRepo.On("FindByID", mock.Anything, expectedTask.ID).Return(expectedTask, nil)

		retrievedTask, err := service.GetByID(ctx, expectedTask.ID)
		require.NoError(t, err)
		assert.Equal(t, expectedTask.ID, retrievedTask.ID)
		assert.Equal(t, createCmd.Title, retrievedTask.Title)
		assert.Equal(t, createCmd.Description, retrievedTask.Description)
		assert.Equal(t, entity.TaskStatusPending, retrievedTask.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("task not found error", func(t *testing.T) {
		mockRepo := testmock.NewTaskRepository()
		taskChannel := make(chan *entity.Task, 10)
		service := NewTaskService(mockRepo, taskChannel)

		testID := uint64(999)
		mockRepo.On("FindByID", mock.Anything, testID).Return(nil, repository.ErrTaskNotFound)

		_, err := service.GetByID(ctx, testID)
		require.Error(t, err)
		
		var appErr *apperror.AppError
		assert.ErrorAs(t, err, &appErr)
		assert.Equal(t, "NOT_FOUND", appErr.Code)
		assert.Contains(t, appErr.Message, "task not found")

		mockRepo.AssertExpectations(t)
	})

	t.Run("failed to get task", func(t *testing.T) {
		mockRepo := testmock.NewTaskRepository()
		taskChannel := make(chan *entity.Task, 10)
		service := NewTaskService(mockRepo, taskChannel)

		dbErr := errors.New("database connection failed")
		mockRepo.On("FindByID", mock.Anything, mock.Anything).Return(nil, dbErr)

		_, err := service.GetByID(ctx, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get task")
		assert.Contains(t, err.Error(), dbErr.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_GetAll(t *testing.T) {
	ctx := context.Background()

	t.Run("successful retrieval of all tasks", func(t *testing.T) {
		mockRepo := testmock.NewTaskRepository()
		taskChannel := make(chan *entity.Task, 10)
		service := NewTaskService(mockRepo, taskChannel)

		expectedTasks := []*entity.Task{
			{ID: 1, Title: "Task 1", Description: "Description 1", Status: entity.TaskStatusPending},
			{ID: 2, Title: "Task 2", Description: "Description 2", Status: entity.TaskStatusPending},
		}
		mockRepo.On("FindAll", mock.Anything).Return(expectedTasks, nil)

		allTasks, err := service.GetAll(ctx)
		require.NoError(t, err)
		assert.Len(t, allTasks, 2)

		mockRepo.AssertExpectations(t)
	})

	t.Run("empty list when no tasks exist", func(t *testing.T) {
		mockRepo := testmock.NewTaskRepository()
		taskChannel := make(chan *entity.Task, 10)
		service := NewTaskService(mockRepo, taskChannel)

		mockRepo.On("FindAll", mock.Anything).Return([]*entity.Task{}, nil)

		allTasks, err := service.GetAll(ctx)
		require.NoError(t, err)
		assert.Empty(t, allTasks)

		mockRepo.AssertExpectations(t)
	})

	t.Run("failed to get tasks", func(t *testing.T) {
		mockRepo := testmock.NewTaskRepository()
		taskChannel := make(chan *entity.Task, 10)
		service := NewTaskService(mockRepo, taskChannel)

		dbErr := errors.New("database connection failed")
		mockRepo.On("FindAll", mock.Anything).Return(nil, dbErr)

		allTasks, err := service.GetAll(ctx)
		assert.Nil(t, allTasks)
		require.Contains(t, err.Error(), "failed to get tasks: "+dbErr.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_ConcurrentCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("multiple concurrent task submissions", func(t *testing.T) {
		mockRepo := testmock.NewTaskRepository()
		taskChannel := make(chan *entity.Task, 100)
		service := NewTaskService(mockRepo, taskChannel)

		const numTasks = 50
		var wg sync.WaitGroup

		// Setup mock to handle all concurrent calls
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Times(numTasks)

		// Create tasks concurrently
		for i := 0; i < numTasks; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				createCmd := &contracts.CreateTask{
					Title:       fmt.Sprintf("Task %d", index),
					Description: fmt.Sprintf("Description for task %d", index),
				}

				if err := service.Create(ctx, createCmd); err != nil {
					t.Errorf("unexpected error during concurrent creation: %v", err)
				}
			}(i)
		}

		wg.Wait()

		for i := 0; i < numTasks; i++ {
			select {
			case task := <-taskChannel:
				assert.Equal(t, entity.TaskStatusPending, task.Status)
			case <-time.After(5 * time.Second):
				t.Fatalf("timeout waiting for task %d", i)
			}
		}

		mockRepo.AssertExpectations(t)
	})
}
