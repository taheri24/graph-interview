package dto

import (
	"taheri24.ir/graph1/internal/types"

	"github.com/google/uuid"
)

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Title       string           `json:"title" binding:"required,min=1,max=200"`
	Description string           `json:"description" binding:"max=1000"`
	Status      types.TaskStatus `json:"status" binding:"omitempty,oneof=pending in_progress completed"`
	Assignee    string           `json:"assignee" binding:"max=100"`
}

// UpdateTaskRequest represents the request body for updating a task
type UpdateTaskRequest struct {
	Title       *string           `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string           `json:"description" binding:"omitempty,max=1000"`
	Status      *types.TaskStatus `json:"status" binding:"omitempty,oneof=pending in_progress completed"`
	Assignee    *string           `json:"assignee" binding:"omitempty,max=100"`
}

// TaskResponse represents the response body for a task
type TaskResponse struct {
	ID          uuid.UUID        `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Status      types.TaskStatus `json:"status"`
	Assignee    string           `json:"assignee"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
}

// TaskListResponse represents the response body for listing tasks
type TaskListResponse struct {
	Tasks       []TaskResponse `json:"tasks"`
	Total       int64          `json:"total"`
	Page        int            `json:"page"`
	Limit       int            `json:"limit"`
	HasNext     bool           `json:"has_next"`
	HasPrevious bool           `json:"has_previous"`
}
