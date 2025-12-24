package task

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"taheri24.ir/graph1/internal/dto"
	"taheri24.ir/graph1/internal/types"
)

func TestValidateCreateTaskRequest(t *testing.T) {
	tests := []struct {
		name     string
		req      dto.CreateTaskRequest
		expected []ValidationError
	}{
		{
			name: "valid request",
			req: dto.CreateTaskRequest{
				Title:       "Test Task",
				Description: "Test Description",
				Status:      types.StatusPending,
				Assignee:    "test@example.com",
			},
			expected: nil,
		},
		{
			name: "missing title",
			req: dto.CreateTaskRequest{
				Description: "Test Description",
				Status:      types.StatusPending,
				Assignee:    "test@example.com",
			},
			expected: []ValidationError{
				{Field: "title", Message: "title is required"},
			},
		},
		{
			name: "empty title",
			req: dto.CreateTaskRequest{
				Title:       "",
				Description: "Test Description",
				Status:      types.StatusPending,
				Assignee:    "test@example.com",
			},
			expected: []ValidationError{
				{Field: "title", Message: "title is required"},
			},
		},
		{
			name: "title too long",
			req: dto.CreateTaskRequest{
				Title:       string(make([]byte, 201)),
				Description: "Test Description",
				Status:      types.StatusPending,
				Assignee:    "test@example.com",
			},
			expected: []ValidationError{
				{Field: "title", Message: "title must be at most 200 characters"},
			},
		},
		{
			name: "description too long",
			req: dto.CreateTaskRequest{
				Title:       "Test Task",
				Description: string(make([]byte, 1001)),
				Status:      types.StatusPending,
				Assignee:    "test@example.com",
			},
			expected: []ValidationError{
				{Field: "description", Message: "description must be at most 1000 characters"},
			},
		},
		{
			name: "invalid status",
			req: dto.CreateTaskRequest{
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
			req: dto.CreateTaskRequest{
				Title:       "Test Task",
				Description: "Test Description",
				Status:      types.StatusPending,
				Assignee:    string(make([]byte, 101)),
			},
			expected: []ValidationError{
				{Field: "assignee", Message: "assignee must be at most 100 characters"},
			},
		},
		{
			name: "multiple errors",
			req: dto.CreateTaskRequest{
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
		{
			name: "nil values",
			req: dto.CreateTaskRequest{
				Title:       "",
				Description: "",
				Status:      "",
				Assignee:    "",
			},
			expected: []ValidationError{
				{Field: "title", Message: "title is required"},
			},
		},
		{
			name: "extremely long strings",
			req: dto.CreateTaskRequest{
				Title:       string(make([]byte, 10000)),
				Description: string(make([]byte, 10000)),
				Status:      types.StatusPending,
				Assignee:    string(make([]byte, 10000)),
			},
			expected: []ValidationError{
				{Field: "title", Message: "title must be at most 200 characters"},
				{Field: "description", Message: "description must be at most 1000 characters"},
				{Field: "assignee", Message: "assignee must be at most 100 characters"},
			},
		},
		{
			name: "special characters in strings",
			req: dto.CreateTaskRequest{
				Title:       "Task with !@#$%^&*()",
				Description: "Description with\nnewlines\tand\ttabs",
				Status:      types.StatusCompleted,
				Assignee:    "user+tag@example.co.uk",
			},
			expected: nil, // should be valid
		},
		{
			name: "unicode characters",
			req: dto.CreateTaskRequest{
				Title:       "创建任务",
				Description: "任务描述 📝",
				Status:      types.StatusInProgress,
				Assignee:    "用户@例子.中国",
			},
			expected: nil, // should be valid
		},
		{
			name: "exactly at limits",
			req: dto.CreateTaskRequest{
				Title:       string(make([]byte, 200)),  // exactly 200 chars
				Description: string(make([]byte, 1000)), // exactly 1000 chars
				Status:      types.StatusPending,
				Assignee:    string(make([]byte, 100)), // exactly 100 chars
			},
			expected: nil, // should be valid
		},
		{
			name: "one character over limits",
			req: dto.CreateTaskRequest{
				Title:       string(make([]byte, 201)),  // 201 chars
				Description: string(make([]byte, 1001)), // 1001 chars
				Status:      types.StatusPending,
				Assignee:    string(make([]byte, 101)), // 101 chars
			},
			expected: []ValidationError{
				{Field: "title", Message: "title must be at most 200 characters"},
				{Field: "description", Message: "description must be at most 1000 characters"},
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
		req      dto.UpdateTaskRequest
		expected []ValidationError
	}{
		{
			name:     "valid request with no updates",
			req:      dto.UpdateTaskRequest{},
			expected: nil,
		},
		{
			name: "valid title update",
			req: dto.UpdateTaskRequest{
				Title: stringPtr("Updated Title"),
			},
			expected: nil,
		},
		{
			name: "empty title update",
			req: dto.UpdateTaskRequest{
				Title: stringPtr(""),
			},
			expected: []ValidationError{
				{Field: "title", Message: "title cannot be empty"},
			},
		},
		{
			name: "title too long",
			req: dto.UpdateTaskRequest{
				Title: stringPtr(string(make([]byte, 201))),
			},
			expected: []ValidationError{
				{Field: "title", Message: "title must be at most 200 characters"},
			},
		},
		{
			name: "description too long",
			req: dto.UpdateTaskRequest{
				Description: stringPtr(string(make([]byte, 1001))),
			},
			expected: []ValidationError{
				{Field: "description", Message: "description must be at most 1000 characters"},
			},
		},
		{
			name: "invalid status",
			req: dto.UpdateTaskRequest{
				Status: statusPtr("invalid"),
			},
			expected: []ValidationError{
				{Field: "status", Message: "status must be one of: pending, in_progress, completed"},
			},
		},
		{
			name: "assignee too long",
			req: dto.UpdateTaskRequest{
				Assignee: stringPtr(string(make([]byte, 101))),
			},
			expected: []ValidationError{
				{Field: "assignee", Message: "assignee must be at most 100 characters"},
			},
		},
		{
			name: "multiple errors",
			req: dto.UpdateTaskRequest{
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
		{
			name: "nil pointers",
			req: dto.UpdateTaskRequest{
				Title:       nil,
				Description: nil,
				Status:      nil,
				Assignee:    nil,
			},
			expected: nil,
		},
		{
			name: "extremely long strings",
			req: dto.UpdateTaskRequest{
				Title:       stringPtr(string(make([]byte, 10000))), // way over limit
				Description: stringPtr(string(make([]byte, 10000))),
				Assignee:    stringPtr(string(make([]byte, 10000))),
			},
			expected: []ValidationError{
				{Field: "title", Message: "title must be at most 200 characters"},
				{Field: "description", Message: "description must be at most 1000 characters"},
				{Field: "assignee", Message: "assignee must be at most 100 characters"},
			},
		},
		{
			name: "special characters in strings",
			req: dto.UpdateTaskRequest{
				Title:       stringPtr("Valid Title!@#$%^&*()"),
				Description: stringPtr("Valid description with\nnewlines\tand\ttabs"),
				Assignee:    stringPtr("user+tag@example.com"),
			},
			expected: nil, // should be valid
		},
		{
			name: "unicode characters",
			req: dto.UpdateTaskRequest{
				Title:       stringPtr("任务标题"),
				Description: stringPtr("任务描述 📝"),
				Assignee:    stringPtr("用户@例子.com"),
			},
			expected: nil, // should be valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateUpdateTaskRequest(tt.req)
			assert.Equal(t, tt.expected, result)
		})
	}
}
