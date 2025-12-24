package models

import (
	"time"

	"taheri24.ir/graph1/internal/middleware"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusInProgress TaskStatus = "in_progress"
	StatusCompleted  TaskStatus = "completed"
)

type Task struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primary_key"`
	Title       string         `json:"title" gorm:"not null"`
	Description string         `json:"description" gorm:"type:text"`
	Status      TaskStatus     `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Assignee    string         `json:"assignee" gorm:"type:varchar(100)"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// BeforeCreate hook to generate UUID for new tasks
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}

	logger := middleware.GetLoggerFromContext(tx.Statement.Context)
	logger.Info("Creating task", "id", t.ID.String(), "title", t.Title)

	return nil
}

// BeforeUpdate hook to log task updates
func (t *Task) BeforeUpdate(tx *gorm.DB) error {
	logger := middleware.GetLoggerFromContext(tx.Statement.Context)
	logger.Info("Updating task", "id", t.ID.String(), "status", string(t.Status), "assignee", t.Assignee)

	return nil
}

// BeforeDelete hook to log task deletions
func (t *Task) BeforeDelete(tx *gorm.DB) error {
	logger := middleware.GetLoggerFromContext(tx.Statement.Context)
	logger.Info("Deleting task", "id", t.ID.String(), "title", t.Title)

	return nil
}

func (Task) TableName() string {
	return "tasks"
}
