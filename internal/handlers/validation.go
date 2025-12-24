package handlers

import (
	"strings"

	"taheri24.ir/graph1/internal/dto"
	"taheri24.ir/graph1/internal/types"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// ValidateCreateTaskRequest validates a CreateTaskRequest
func ValidateCreateTaskRequest(req dto.CreateTaskRequest) []ValidationError {
	var errors []ValidationError

	// Validate Title
	if strings.TrimSpace(req.Title) == "" {
		errors = append(errors, ValidationError{Field: "title", Message: "title is required"})
	} else if len(req.Title) < 1 {
		errors = append(errors, ValidationError{Field: "title", Message: "title must be at least 1 character"})
	} else if len(req.Title) > 200 {
		errors = append(errors, ValidationError{Field: "title", Message: "title must be at most 200 characters"})
	}

	// Validate Description
	if len(req.Description) > 1000 {
		errors = append(errors, ValidationError{Field: "description", Message: "description must be at most 1000 characters"})
	}

	// Validate Status
	if req.Status != "" && !isValidTaskStatus(req.Status) {
		errors = append(errors, ValidationError{Field: "status", Message: "status must be one of: pending, in_progress, completed"})
	}

	// Validate Assignee
	if len(req.Assignee) > 100 {
		errors = append(errors, ValidationError{Field: "assignee", Message: "assignee must be at most 100 characters"})
	}

	return errors
}

// ValidateUpdateTaskRequest validates an UpdateTaskRequest
func ValidateUpdateTaskRequest(req dto.UpdateTaskRequest) []ValidationError {
	var errors []ValidationError

	// Validate Title
	if req.Title != nil {
		title := *req.Title
		if strings.TrimSpace(title) == "" {
			errors = append(errors, ValidationError{Field: "title", Message: "title cannot be empty"})
		} else if len(title) < 1 {
			errors = append(errors, ValidationError{Field: "title", Message: "title must be at least 1 character"})
		} else if len(title) > 200 {
			errors = append(errors, ValidationError{Field: "title", Message: "title must be at most 200 characters"})
		}
	}

	// Validate Description
	if req.Description != nil && len(*req.Description) > 1000 {
		errors = append(errors, ValidationError{Field: "description", Message: "description must be at most 1000 characters"})
	}

	// Validate Status
	if req.Status != nil && !isValidTaskStatus(*req.Status) {
		errors = append(errors, ValidationError{Field: "status", Message: "status must be one of: pending, in_progress, completed"})
	}

	// Validate Assignee
	if req.Assignee != nil && len(*req.Assignee) > 100 {
		errors = append(errors, ValidationError{Field: "assignee", Message: "assignee must be at most 100 characters"})
	}

	return errors
}

// isValidTaskStatus checks if the status is valid
func isValidTaskStatus(status types.TaskStatus) bool {
	switch status {
	case types.StatusPending, types.StatusInProgress, types.StatusCompleted:
		return true
	default:
		return false
	}
}
