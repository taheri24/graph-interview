package models_test

import (
	"testing"
	"time"

	"taheri24.ir/graph1/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

func TestTaskBeforeCreateHook(t *testing.T) {
	// Create a mock database connection
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Test case 1: Task with nil ID should get a new UUID generated
	task := models.Task{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      models.StatusPending,
		Assignee:    "test@example.com",
	}
	assert.Equal(t, uuid.Nil, task.ID)

	// Call BeforeCreate hook
	err = task.BeforeCreate(gormDB)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, task.ID)

	// Verify it's a valid UUID
	assert.NotEqual(t, uuid.Nil, task.ID)

	// Test case 2: Task with existing ID should preserve the ID
	existingID := uuid.New()
	task2 := models.Task{
		ID:          existingID,
		Title:       "Test Task 2",
		Description: "Test Description 2",
		Status:      models.StatusInProgress,
		Assignee:    "test2@example.com",
	}

	err = task2.BeforeCreate(gormDB)
	assert.NoError(t, err)
	assert.Equal(t, existingID, task2.ID)

	// Clean up
	assert.NoError(t, mock.ExpectationsWereMet())
}
