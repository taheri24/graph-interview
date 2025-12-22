package models_test

import (
	"testing"
	"time"

	"taheri24.ir/graph1/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTaskTableName(t *testing.T) {
	task := models.Task{}
	assert.Equal(t, "tasks", task.TableName())
}

func TestTaskStatusConstants(t *testing.T) {
	assert.Equal(t, models.TaskStatus("pending"), models.StatusPending)
	assert.Equal(t, models.TaskStatus("in_progress"), models.StatusInProgress)
	assert.Equal(t, models.TaskStatus("completed"), models.StatusCompleted)
}

func TestTaskModel(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	task := models.Task{
		ID:          id,
		Title:       "Test Title",
		Description: "Test Description",
		Status:      models.StatusPending,
		Assignee:    "test@example.com",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assert.Equal(t, id, task.ID)
	assert.Equal(t, "Test Title", task.Title)
	assert.Equal(t, "Test Description", task.Description)
	assert.Equal(t, models.StatusPending, task.Status)
	assert.Equal(t, "test@example.com", task.Assignee)
	assert.Equal(t, now, task.CreatedAt)
	assert.Equal(t, now, task.UpdatedAt)
}

func TestTaskModelWithDefaultValues(t *testing.T) {
	task := models.Task{
		ID:        uuid.New(),
		Title:     "Test Task",
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test that empty string fields are handled correctly
	assert.Equal(t, "", task.Description)
	assert.Equal(t, "", task.Assignee)
}

func TestTaskStatusValidation(t *testing.T) {
	// Test valid status values
	validStatuses := []models.TaskStatus{
		models.StatusPending,
		models.StatusInProgress,
		models.StatusCompleted,
	}

	for _, status := range validStatuses {
		assert.NotEmpty(t, string(status))
	}

	// Test custom status
	customStatus := models.TaskStatus("custom")
	assert.Equal(t, models.TaskStatus("custom"), customStatus)
}

func TestTaskID(t *testing.T) {
	task := models.Task{}

	// Test with nil UUID
	assert.Equal(t, uuid.Nil, task.ID)

	// Test with valid UUID
	id := uuid.New()
	task.ID = id
	assert.Equal(t, id, task.ID)
	assert.NotEqual(t, uuid.Nil, task.ID)
}
