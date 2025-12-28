package entity

import (
	"time"
)

type Task struct {
	ID          string `gorm:"primaryKey"`
	Title       string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewTask(title, description string) *Task {
	return &Task{
		Title:       title,
		Description: description,
	}
}

func (Task) TableName() string {
	return "tasks"
}
