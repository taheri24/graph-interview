package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"taheri24.ir/graph1/internal/cache"
	"taheri24.ir/graph1/internal/models"
)

// MockTaskRepository for testing
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) GetByID(id uuid.UUID) (*models.Task, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Task), args.Error(1)
}

func (m *MockTaskRepository) GetAll(page, limit int, status, assignee string) ([]models.Task, int64, error) {
	args := m.Called(page, limit, status, assignee)
	return args.Get(0).([]models.Task), args.Get(1).(int64), args.Error(2)
}

func (m *MockTaskRepository) Update(task *models.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTaskRepository) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTaskRepository) Health() error {
	args := m.Called()
	return args.Error(0)
}

// MockCache for testing
type MockCache[T any] struct {
	mock.Mock
}

// Get implements [cache.CacheInterface].
func (m *MockCache[T]) Get(id string) (T, error) {
	args := m.Called(id)
	return args.Get(0).(T), args.Error(1)
}

// GetAll implements [cache.CacheInterface].
func (m *MockCache[T]) GetAll() ([]T, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]T), args.Error(1)
}

// Invalidate implements [cache.CacheInterface].
func (m *MockCache[T]) Invalidate(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// InvalidateAll implements [cache.CacheInterface].
func (m *MockCache[T]) InvalidateAll() error {
	args := m.Called()
	return args.Error(0)
}

// Set implements [cache.CacheInterface].
func (m *MockCache[T]) Set(id string, item T) error {
	args := m.Called(id, item)
	return args.Error(0)
}

// SetAll implements [cache.CacheInterface].
func (m *MockCache[T]) SetAll(tasks []T) error {
	args := m.Called(tasks)
	return args.Error(0)
}

var _ cache.CacheInterface[any] = (*MockCache[any])(nil)

// setupTestRouter creates a new Gin router for testing
func setupTestRouter() *gin.Engine {
	return gin.New()
}

func TestNewCachedTaskHandler(t *testing.T) {
	mockRepo := &MockTaskRepository{}
	mockCache := &MockCache[models.Task]{}

	handler := NewCachedTaskHandler(mockRepo, mockCache)

	assert.NotNil(t, handler)
	assert.Equal(t, mockRepo, handler.repo)
	assert.Equal(t, mockCache, handler.cache)
}

func TestCachedTaskHandler_GetTasks_CacheHit(t *testing.T) {
	mockRepo := &MockTaskRepository{}
	mockCache := &MockCache[models.Task]{}
	handler := NewCachedTaskHandler(mockRepo, mockCache)

	// Setup cached tasks
	cachedTasks := []models.Task{
		{
			ID:          uuid.New(),
			Title:       "Cached Task 1",
			Description: "Cached Description 1",
			Status:      models.StatusPending,
			Assignee:    "Cached User 1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Title:       "Cached Task 2",
			Description: "Cached Description 2",
			Status:      models.StatusInProgress,
			Assignee:    "Cached User 2",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	mockCache.On("GetAll").Return(cachedTasks, nil)

	router := setupTestRouter()
	router.GET("/tasks", handler.GetTasks)

	req, _ := http.NewRequest("GET", "/tasks", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response TaskListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Tasks, 2)
	assert.Equal(t, int64(2), response.Total)

	mockCache.AssertExpectations(t)
}

func TestCachedTaskHandler_GetTasks_CacheMiss(t *testing.T) {
	mockRepo := &MockTaskRepository{}
	mockCache := &MockCache[models.Task]{}
	handler := NewCachedTaskHandler(mockRepo, mockCache)

	// Setup database response
	dbTasks := []models.Task{
		{
			ID:          uuid.New(),
			Title:       "DB Task 1",
			Description: "DB Description 1",
			Status:      models.StatusPending,
			Assignee:    "DB User 1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	mockCache.On("GetAll").Return(nil, nil) // Cache miss - nil slice
	mockCache.On("SetAll", mock.AnythingOfType("[]models.Task")).Return(nil)
	mockRepo.On("GetAll", 1, 10, "", "").Return(dbTasks, int64(1), nil)

	router := setupTestRouter()
	router.GET("/tasks", handler.GetTasks)

	req, _ := http.NewRequest("GET", "/tasks", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response TaskListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Len(t, response.Tasks, 1)
	assert.Equal(t, int64(1), response.Total)

	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestCachedTaskHandler_GetTasks_CacheError(t *testing.T) {
	mockRepo := &MockTaskRepository{}
	mockCache := &MockCache[models.Task]{}
	handler := NewCachedTaskHandler(mockRepo, mockCache)

	mockCache.On("GetAll").Return([]models.Task(nil), assert.AnError)

	router := setupTestRouter()
	router.GET("/tasks", handler.GetTasks)

	req, _ := http.NewRequest("GET", "/tasks", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Cache error", response.Error)

	mockCache.AssertExpectations(t)
}

func TestCachedTaskHandler_GetTask_CacheHit(t *testing.T) {
	mockRepo := &MockTaskRepository{}
	mockCache := &MockCache[models.Task]{}
	handler := NewCachedTaskHandler(mockRepo, mockCache)

	taskID := uuid.New()
	cachedTask := &models.Task{
		ID:          taskID,
		Title:       "Cached Task",
		Description: "Cached Description",
		Status:      models.StatusPending,
		Assignee:    "Cached User",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockCache.On("Get", taskID.String()).Return(*cachedTask, nil)

	router := setupTestRouter()
	router.GET("/tasks/:id", handler.GetTask)

	req, _ := http.NewRequest("GET", "/tasks/"+taskID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response TaskResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, taskID, response.ID)
	assert.Equal(t, "Cached Task", response.Title)

	mockCache.AssertExpectations(t)
}

func TestCachedTaskHandler_GetTask_CacheMiss(t *testing.T) {
	mockRepo := &MockTaskRepository{}
	mockCache := &MockCache[models.Task]{}
	handler := NewCachedTaskHandler(mockRepo, mockCache)

	taskID := uuid.New()
	dbTask := &models.Task{
		ID:          taskID,
		Title:       "DB Task",
		Description: "DB Description",
		Status:      models.StatusPending,
		Assignee:    "DB User",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockCache.On("Get", taskID.String()).Return(models.Task{}, nil) // Cache miss
	mockCache.On("Set", taskID.String(), *dbTask).Return(nil)
	mockRepo.On("GetByID", taskID).Return(dbTask, nil)

	router := setupTestRouter()
	router.GET("/tasks/:id", handler.GetTask)

	req, _ := http.NewRequest("GET", "/tasks/"+taskID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response TaskResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, taskID, response.ID)
	assert.Equal(t, "DB Task", response.Title)

	mockCache.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestCachedTaskHandler_GetTask_InvalidID(t *testing.T) {
	mockRepo := &MockTaskRepository{}
	mockCache := &MockCache[models.Task]{}
	handler := NewCachedTaskHandler(mockRepo, mockCache)

	router := setupTestRouter()
	router.GET("/tasks/:id", handler.GetTask)

	req, _ := http.NewRequest("GET", "/tasks/invalid-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Invalid task ID", response.Error)
}

func TestCachedTaskHandler_CreateTask(t *testing.T) {
	mockRepo := &MockTaskRepository{}
	mockCache := &MockCache[models.Task]{}
	handler := NewCachedTaskHandler(mockRepo, mockCache)

	createReq := CreateTaskRequest{
		Title:       "New Task",
		Description: "New Description",
		Status:      models.StatusPending,
		Assignee:    "New User",
	}

	mockRepo.On("Create", mock.AnythingOfType("*models.Task")).Return(nil)
	mockCache.On("InvalidateAll").Return(nil)
	mockRepo.On("GetAll", 1, 1, "", "").Return([]models.Task{}, int64(1), nil)

	router := setupTestRouter()
	router.POST("/tasks", handler.CreateTask)

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response TaskResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "New Task", response.Title)
	assert.Equal(t, models.StatusPending, response.Status)

	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestCachedTaskHandler_UpdateTask(t *testing.T) {
	mockRepo := &MockTaskRepository{}
	mockCache := &MockCache[models.Task]{}
	handler := NewCachedTaskHandler(mockRepo, mockCache)

	taskID := uuid.New()
	existingTask := &models.Task{
		ID:          taskID,
		Title:       "Old Title",
		Description: "Old Description",
		Status:      models.StatusPending,
		Assignee:    "Old User",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	updateReq := UpdateTaskRequest{
		Title:       stringPtr("Updated Title"),
		Description: stringPtr("Updated Description"),
		Status:      func() *models.TaskStatus { s := models.StatusInProgress; return &s }(),
		Assignee:    stringPtr("Updated User"),
	}

	mockRepo.On("GetByID", taskID).Return(existingTask, nil)
	mockRepo.On("Update", mock.AnythingOfType("*models.Task")).Return(nil)
	mockCache.On("InvalidateAll").Return(nil)
	mockCache.On("Invalidate", taskID.String()).Return(nil)

	router := setupTestRouter()
	router.PUT("/tasks/:id", handler.UpdateTask)

	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/tasks/"+taskID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response TaskResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, taskID, response.ID)
	assert.Equal(t, "Updated Title", response.Title)
	assert.Equal(t, models.StatusInProgress, response.Status)

	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestCachedTaskHandler_DeleteTask(t *testing.T) {
	mockRepo := &MockTaskRepository{}
	mockCache := &MockCache[models.Task]{}
	handler := NewCachedTaskHandler(mockRepo, mockCache)

	taskID := uuid.New()

	mockRepo.On("Delete", taskID).Return(nil)
	mockCache.On("InvalidateAll").Return(nil)
	mockCache.On("Invalidate", taskID.String()).Return(nil)
	mockRepo.On("GetAll", 1, 1, "", "").Return([]models.Task{}, int64(0), nil)

	router := setupTestRouter()
	router.DELETE("/tasks/:id", handler.DeleteTask)

	req, _ := http.NewRequest("DELETE", "/tasks/"+taskID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())

	mockRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
