package mock

import (
	"context"
	"task-pool/internal/domain/entity"

	"github.com/stretchr/testify/mock"
)

// TaskRepository is a mock implementation of TaskRepository for testing using testify/mock
type TaskRepository struct {
	mock.Mock
}

// NewTaskRepository creates a new instance of TaskRepository
func NewTaskRepository() *TaskRepository {
	return &TaskRepository{}
}

func (m *TaskRepository) Create(ctx context.Context, task *entity.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *TaskRepository) FindByID(ctx context.Context, id uint64) (*entity.Task, error) {
	args := m.Called(ctx, id)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*entity.Task), args.Error(1)
}

func (m *TaskRepository) FindAll(ctx context.Context) ([]*entity.Task, error) {
	args := m.Called(ctx)

	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*entity.Task), args.Error(1)
}

func (m *TaskRepository) Update(ctx context.Context, task *entity.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}
