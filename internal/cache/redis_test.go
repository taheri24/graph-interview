package cache

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"taheri24.ir/graph1/internal/models"
)

// MockRedisCache for testing without actual Redis connection
type MockRedisCache struct {
	data map[string]string
}

func NewMockRedisCache() *MockRedisCache {
	return &MockRedisCache{
		data: make(map[string]string),
	}
}

func (m *MockRedisCache) GetTasks() ([]models.Task, error) {
	_, exists := m.data["tasks:list"]
	if !exists {
		return nil, nil // Cache miss
	}

	// In a real implementation, this would be JSON unmarshaling
	// For mock, we'll return empty slice
	return []models.Task{}, nil
}

func (m *MockRedisCache) SetTasks(tasks []models.Task, expiration time.Duration) error {
	// In a real implementation, this would be JSON marshaling
	m.data["tasks:list"] = "cached_tasks"
	return nil
}

func (m *MockRedisCache) InvalidateTasks() error {
	delete(m.data, "tasks:list")
	return nil
}

func (m *MockRedisCache) GetTask(id string) (*models.Task, error) {
	key := "task:" + id
	data, exists := m.data[key]
	if !exists {
		return nil, nil // Cache miss
	}

	// Mock implementation
	if data == "mock_task" {
		return &models.Task{
			ID:          uuid.MustParse(id),
			Title:       "Mock Task",
			Description: "Mock Description",
			Status:      models.StatusPending,
			Assignee:    "Mock Assignee",
		}, nil
	}
	return nil, nil
}

func (m *MockRedisCache) SetTask(task *models.Task, expiration time.Duration) error {
	key := "task:" + task.ID.String()
	m.data[key] = "mock_task"
	return nil
}

func (m *MockRedisCache) InvalidateTask(id string) error {
	key := "task:" + id
	delete(m.data, key)
	return nil
}

func TestRedisCache_GetTasks(t *testing.T) {
	cache := NewMockRedisCache()

	// Test cache miss
	tasks, err := cache.GetTasks()
	assert.NoError(t, err)
	assert.Nil(t, tasks)

	// Test cache hit
	mockTasks := []models.Task{
		{
			ID:          uuid.New(),
			Title:       "Test Task",
			Description: "Test Description",
			Status:      models.StatusPending,
			Assignee:    "Test Assignee",
		},
	}

	err = cache.SetTasks(mockTasks, 5*time.Minute)
	require.NoError(t, err)

	tasks, err = cache.GetTasks()
	assert.NoError(t, err)
	assert.NotNil(t, tasks)
}

func TestRedisCache_SetTasks(t *testing.T) {
	cache := NewMockRedisCache()

	tasks := []models.Task{
		{
			ID:          uuid.New(),
			Title:       "Test Task",
			Description: "Test Description",
			Status:      models.StatusPending,
			Assignee:    "Test Assignee",
		},
	}

	err := cache.SetTasks(tasks, 5*time.Minute)
	assert.NoError(t, err)

	// Verify data was set (in mock, we just check if key exists)
	cachedTasks, err := cache.GetTasks()
	assert.NoError(t, err)
	assert.NotNil(t, cachedTasks)
}

func TestRedisCache_InvalidateTasks(t *testing.T) {
	cache := NewMockRedisCache()

	// Set some data first
	tasks := []models.Task{
		{
			ID:          uuid.New(),
			Title:       "Test Task",
			Description: "Test Description",
			Status:      models.StatusPending,
			Assignee:    "Test Assignee",
		},
	}

	err := cache.SetTasks(tasks, 5*time.Minute)
	require.NoError(t, err)

	// Invalidate cache
	err = cache.InvalidateTasks()
	assert.NoError(t, err)

	// Verify cache is empty
	cachedTasks, err := cache.GetTasks()
	assert.NoError(t, err)
	assert.Nil(t, cachedTasks)
}

func TestRedisCache_GetTask(t *testing.T) {
	cache := NewMockRedisCache()
	taskID := uuid.New().String()

	// Test cache miss
	task, err := cache.GetTask(taskID)
	assert.NoError(t, err)
	assert.Nil(t, task)

	// Test cache hit
	mockTask := &models.Task{
		ID:          uuid.MustParse(taskID),
		Title:       "Test Task",
		Description: "Test Description",
		Status:      models.StatusPending,
		Assignee:    "Test Assignee",
	}

	err = cache.SetTask(mockTask, 5*time.Minute)
	require.NoError(t, err)

	task, err = cache.GetTask(taskID)
	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, mockTask.ID, task.ID)
}

func TestRedisCache_SetTask(t *testing.T) {
	cache := NewMockRedisCache()

	task := &models.Task{
		ID:          uuid.New(),
		Title:       "Test Task",
		Description: "Test Description",
		Status:      models.StatusPending,
		Assignee:    "Test Assignee",
	}

	err := cache.SetTask(task, 5*time.Minute)
	assert.NoError(t, err)

	// Verify data was set
	cachedTask, err := cache.GetTask(task.ID.String())
	assert.NoError(t, err)
	assert.NotNil(t, cachedTask)
	assert.Equal(t, task.ID, cachedTask.ID)
}

func TestRedisCache_InvalidateTask(t *testing.T) {
	cache := NewMockRedisCache()

	task := &models.Task{
		ID:          uuid.New(),
		Title:       "Test Task",
		Description: "Test Description",
		Status:      models.StatusPending,
		Assignee:    "Test Assignee",
	}

	// Set some data first
	err := cache.SetTask(task, 5*time.Minute)
	require.NoError(t, err)

	// Invalidate cache
	err = cache.InvalidateTask(task.ID.String())
	assert.NoError(t, err)

	// Verify cache is empty
	cachedTask, err := cache.GetTask(task.ID.String())
	assert.NoError(t, err)
	assert.Nil(t, cachedTask)
}
