package models

import (
	"time"

	"taheri24.ir/graph1/internal/types"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Task struct {
	ID          uuid.UUID        `json:"id" gorm:"type:uuid;primary_key"`
	Title       string           `json:"title" gorm:"not null"`
	Description string           `json:"description" gorm:"type:text"`
	Status      types.TaskStatus `json:"status" gorm:"type:varchar(20);default:'pending'"`
	Assignee    string           `json:"assignee" gorm:"type:varchar(100)"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	DeletedAt   gorm.DeletedAt   `json:"-" gorm:"index"`
}

func (Task) TableName() string {
	return "tasks"
}

func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}
