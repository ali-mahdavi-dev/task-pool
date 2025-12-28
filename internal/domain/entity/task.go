package entity

import (
	"time"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID          uint64 `gorm:"primaryKey"`
	Title       string
	Description string
	Status      TaskStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewTask(title, description string, status TaskStatus) *Task {
	return &Task{
		Title:       title,
		Status:      status,
		Description: description,
	}
}

func (Task) TableName() string {
	return "tasks"
}

func (t *Task) Complete() {
	t.Status = TaskStatusCompleted
}

func (t *Task) Failed() {
	t.Status = TaskStatusFailed
}
