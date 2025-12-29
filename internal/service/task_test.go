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

type testFixture struct {
	mockRepo    *testmock.TaskRepository
	taskChannel chan *entity.Task
	service     contracts.TaskService
	ctx         context.Context
}

func setupFixture(channelSize ...int) *testFixture {
	size := 10
	if len(channelSize) > 0 && channelSize[0] > 0 {
		size = channelSize[0]
	}

	mockRepo := testmock.NewTaskRepository()
	taskChannel := make(chan *entity.Task, size)
	service := NewTaskService(mockRepo, taskChannel)

	return &testFixture{
		mockRepo:    mockRepo,
		taskChannel: taskChannel,
		service:     service,
		ctx:         context.Background(),
	}
}

func TestTaskService_Create(t *testing.T) {
	t.Run("successful task creation", func(t *testing.T) {
		fixture := setupFixture()

		createCmd := &contracts.CreateTask{
			Title:       "Test Task",
			Description: "Test Description",
		}

		fixture.mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

		err := fixture.service.Create(fixture.ctx, createCmd)
		require.NoError(t, err)

		select {
		case task := <-fixture.taskChannel:
			assert.Equal(t, createCmd.Title, task.Title)
			assert.Equal(t, createCmd.Description, task.Description)
			assert.Equal(t, entity.TaskStatusPending, task.Status)
		case <-time.After(1 * time.Second):
			t.Fatal("task was not sent to channel")
		}

		fixture.mockRepo.AssertExpectations(t)
	})

	t.Run("repository error on create", func(t *testing.T) {
		fixture := setupFixture()

		createCmd := &contracts.CreateTask{
			Title:       "Test Task",
			Description: "Test Description",
		}

		expectedError := errors.New("database connection failed")
		fixture.mockRepo.On("Create", mock.Anything, mock.Anything).Return(expectedError)

		err := fixture.service.Create(fixture.ctx, createCmd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create task")
		assert.Contains(t, err.Error(), "database connection failed")

		// Verify task was not sent to channel
		select {
		case <-fixture.taskChannel:
			t.Fatal("task should not be sent to channel when repository fails")
		default:
			// Expected: channel should be empty
		}

		fixture.mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_GetByID(t *testing.T) {
	t.Run("successful task retrieval by id", func(t *testing.T) {
		fixture := setupFixture()

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
		fixture.mockRepo.On("FindByID", mock.Anything, expectedTask.ID).Return(expectedTask, nil)

		retrievedTask, err := fixture.service.GetByID(fixture.ctx, expectedTask.ID)
		require.NoError(t, err)
		assert.Equal(t, expectedTask.ID, retrievedTask.ID)
		assert.Equal(t, createCmd.Title, retrievedTask.Title)
		assert.Equal(t, createCmd.Description, retrievedTask.Description)
		assert.Equal(t, entity.TaskStatusPending, retrievedTask.Status)

		fixture.mockRepo.AssertExpectations(t)
	})

	t.Run("task not found error", func(t *testing.T) {
		fixture := setupFixture()

		testID := uint64(999)
		fixture.mockRepo.On("FindByID", mock.Anything, testID).Return(nil, repository.ErrTaskNotFound)

		_, err := fixture.service.GetByID(fixture.ctx, testID)
		require.Error(t, err)
		
		var appErr *apperror.AppError
		assert.ErrorAs(t, err, &appErr)
		assert.Equal(t, "NOT_FOUND", appErr.Code)
		assert.Contains(t, appErr.Message, "task not found")

		fixture.mockRepo.AssertExpectations(t)
	})

	t.Run("failed to get task", func(t *testing.T) {
		fixture := setupFixture()

		dbErr := errors.New("database connection failed")
		fixture.mockRepo.On("FindByID", mock.Anything, mock.Anything).Return(nil, dbErr)

		_, err := fixture.service.GetByID(fixture.ctx, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get task")
		assert.Contains(t, err.Error(), dbErr.Error())

		fixture.mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_GetAll(t *testing.T) {
	t.Run("successful retrieval of all tasks", func(t *testing.T) {
		fixture := setupFixture()

		expectedTasks := []*entity.Task{
			{ID: 1, Title: "Task 1", Description: "Description 1", Status: entity.TaskStatusPending},
			{ID: 2, Title: "Task 2", Description: "Description 2", Status: entity.TaskStatusPending},
		}
		fixture.mockRepo.On("FindAll", mock.Anything).Return(expectedTasks, nil)

		allTasks, err := fixture.service.GetAll(fixture.ctx)
		require.NoError(t, err)
		assert.Len(t, allTasks, 2)

		fixture.mockRepo.AssertExpectations(t)
	})

	t.Run("empty list when no tasks exist", func(t *testing.T) {
		fixture := setupFixture()

		fixture.mockRepo.On("FindAll", mock.Anything).Return([]*entity.Task{}, nil)

		allTasks, err := fixture.service.GetAll(fixture.ctx)
		require.NoError(t, err)
		assert.Empty(t, allTasks)

		fixture.mockRepo.AssertExpectations(t)
	})

	t.Run("failed to get tasks", func(t *testing.T) {
		fixture := setupFixture()

		dbErr := errors.New("database connection failed")
		fixture.mockRepo.On("FindAll", mock.Anything).Return(nil, dbErr)

		allTasks, err := fixture.service.GetAll(fixture.ctx)
		assert.Nil(t, allTasks)
		require.Contains(t, err.Error(), "failed to get tasks: "+dbErr.Error())

		fixture.mockRepo.AssertExpectations(t)
	})
}

func TestTaskService_ConcurrentCreate(t *testing.T) {
	t.Run("multiple concurrent task submissions", func(t *testing.T) {
		fixture := setupFixture(100)

		const numTasks = 50
		var wg sync.WaitGroup

		// Setup mock to handle all concurrent calls
		fixture.mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil).Times(numTasks)

		// Create tasks concurrently
		for i := 0; i < numTasks; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				createCmd := &contracts.CreateTask{
					Title:       fmt.Sprintf("Task %d", index),
					Description: fmt.Sprintf("Description for task %d", index),
				}

				if err := fixture.service.Create(fixture.ctx, createCmd); err != nil {
					t.Errorf("unexpected error during concurrent creation: %v", err)
				}
			}(i)
		}

		wg.Wait()

		for i := 0; i < numTasks; i++ {
			select {
			case task := <-fixture.taskChannel:
				assert.Equal(t, entity.TaskStatusPending, task.Status)
			case <-time.After(5 * time.Second):
				t.Fatalf("timeout waiting for task %d", i)
			}
		}

		fixture.mockRepo.AssertExpectations(t)
	})
}
