package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"taheri24.ir/graph1/internal/models"
)

func TestValidateCreateTaskRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      CreateTaskRequest
		expected []ValidationError
	}{
		{
			name: "valid request",
			req: CreateTaskRequest{
				Title:       "Test Task",
				Description: "Test Description",
				Status:      models.StatusPending,
				Assignee:    "test@example.com",
			},
			expected: nil,
		},
		{
			name: "missing title",
			req: CreateTaskRequest{
				Description: "Test Description",
				Status:      models.StatusPending,
				Assignee:    "test@example.com",
			},
			expected: []ValidationError{
				{Field: "title", Message: "title is required"},
			},
		},
		{
			name: "empty title",
			req: CreateTaskRequest{
				Title:       "",
				Description: "Test Description",
				Status:      models.StatusPending,
				Assignee:    "test@example.com",
			},
			expected: []ValidationError{
				{Field: "title", Message: "title is required"},
			},
		},
		{
			name: "title too long",
			req: CreateTaskRequest{
				Title:       string(make([]byte, 201)),
				Description: "Test Description",
				Status:      models.StatusPending,
				Assignee:    "test@example.com",
			},
			expected: []ValidationError{
				{Field: "title", Message: "title must be at most 200 characters"},
			},
		},
		{
			name: "description too long",
			req: CreateTaskRequest{
				Title:       "Test Task",
				Description: string(make([]byte, 1001)),
				Status:      models.StatusPending,
				Assignee:    "test@example.com",
			},
			expected: []ValidationError{
				{Field: "description", Message: "description must be at most 1000 characters"},
			},
		},
		{
			name: "invalid status",
			req: CreateTaskRequest{
				Title:       "Test Task",
				Description: "Test Description",
				Status:      "invalid",
				Assignee:    "test@example.com",
			},
			expected: []ValidationError{
				{Field: "status", Message: "status must be one of: pending, in_progress, completed"},
			},
		},
		{
			name: "assignee too long",
			req: CreateTaskRequest{
				Title:       "Test Task",
				Description: "Test Description",
				Status:      models.StatusPending,
				Assignee:    string(make([]byte, 101)),
			},
			expected: []ValidationError{
				{Field: "assignee", Message: "assignee must be at most 100 characters"},
			},
		},
		{
			name: "multiple errors",
			req: CreateTaskRequest{
				Title:       "",
				Description: string(make([]byte, 1001)),
				Status:      "invalid",
				Assignee:    string(make([]byte, 101)),
			},
			expected: []ValidationError{
				{Field: "title", Message: "title is required"},
				{Field: "description", Message: "description must be at most 1000 characters"},
				{Field: "status", Message: "status must be one of: pending, in_progress, completed"},
				{Field: "assignee", Message: "assignee must be at most 100 characters"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateCreateTaskRequest(tt.req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateUpdateTaskRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      UpdateTaskRequest
		expected []ValidationError
	}{
		{
			name:     "valid request with no updates",
			req:      UpdateTaskRequest{},
			expected: nil,
		},
		{
			name: "valid title update",
			req: UpdateTaskRequest{
				Title: stringPtr("Updated Title"),
			},
			expected: nil,
		},
		{
			name: "empty title update",
			req: UpdateTaskRequest{
				Title: stringPtr(""),
			},
			expected: []ValidationError{
				{Field: "title", Message: "title cannot be empty"},
			},
		},
		{
			name: "title too long",
			req: UpdateTaskRequest{
				Title: stringPtr(string(make([]byte, 201))),
			},
			expected: []ValidationError{
				{Field: "title", Message: "title must be at most 200 characters"},
			},
		},
		{
			name: "description too long",
			req: UpdateTaskRequest{
				Description: stringPtr(string(make([]byte, 1001))),
			},
			expected: []ValidationError{
				{Field: "description", Message: "description must be at most 1000 characters"},
			},
		},
		{
			name: "invalid status",
			req: UpdateTaskRequest{
				Status: statusPtr("invalid"),
			},
			expected: []ValidationError{
				{Field: "status", Message: "status must be one of: pending, in_progress, completed"},
			},
		},
		{
			name: "assignee too long",
			req: UpdateTaskRequest{
				Assignee: stringPtr(string(make([]byte, 101))),
			},
			expected: []ValidationError{
				{Field: "assignee", Message: "assignee must be at most 100 characters"},
			},
		},
		{
			name: "multiple errors",
			req: UpdateTaskRequest{
				Title:       stringPtr(""),
				Description: stringPtr(string(make([]byte, 1001))),
				Status:      statusPtr("invalid"),
				Assignee:    stringPtr(string(make([]byte, 101))),
			},
			expected: []ValidationError{
				{Field: "title", Message: "title cannot be empty"},
				{Field: "description", Message: "description must be at most 1000 characters"},
				{Field: "status", Message: "status must be one of: pending, in_progress, completed"},
				{Field: "assignee", Message: "assignee must be at most 100 characters"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateUpdateTaskRequest(tt.req)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions for creating pointers
func stringPtr(s string) *string {
	return &s
}

func statusPtr(s models.TaskStatus) *models.TaskStatus {
	return &s
}
