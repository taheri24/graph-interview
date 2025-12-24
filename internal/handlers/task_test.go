package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"taheri24.ir/graph1/internal/handlers"
	"taheri24.ir/graph1/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MockTaskRepository implements TaskRepository for testing
type MockTaskRepository struct {
	CreateFunc  func(task *models.Task) error
	GetByIDFunc func(id uuid.UUID) (*models.Task, error)
	GetAllFunc  func(page, limit int, status, assignee string) ([]models.Task, int64, error)
	UpdateFunc  func(task *models.Task) error
	DeleteFunc  func(id uuid.UUID) error
}

// MockCache implements CacheInterface for testing
type MockCache struct {
	GetFunc        func(id string) (models.Task, error)
	SetFunc        func(id string, item models.Task) error
	InvalidateFunc func(id string) error
}

func (m *MockCache) Get(id string) (models.Task, error) {
	if m.GetFunc != nil {
		return m.GetFunc(id)
	}
	return models.Task{}, fmt.Errorf("cache miss")
}

func (m *MockCache) Set(id string, item models.Task) error {
	if m.SetFunc != nil {
		return m.SetFunc(id, item)
	}
	return nil
}

func (m *MockCache) Invalidate(id string) error {
	if m.InvalidateFunc != nil {
		return m.InvalidateFunc(id)
	}
	return nil
}

func (m *MockCache) GetAll() ([]models.Task, error) {
	return nil, nil
}

func (m *MockCache) SetAll(items []models.Task) error {
	return nil
}

func (m *MockCache) InvalidateAll() error {
	return nil
}

func (m *MockTaskRepository) Create(task *models.Task) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(task)
	}
	return nil
}

func (m *MockTaskRepository) GetByID(id uuid.UUID) (*models.Task, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return nil, nil
}

func (m *MockTaskRepository) GetAll(page, limit int, status, assignee string) ([]models.Task, int64, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(page, limit, status, assignee)
	}
	return nil, 0, nil
}

func (m *MockTaskRepository) Update(task *models.Task) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(task)
	}
	return nil
}

func (m *MockTaskRepository) Delete(id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

type TaskHandlerTestSuite struct {
	suite.Suite
	mockRepo  *MockTaskRepository
	mockCache *MockCache
	handler   *handlers.TaskHandler
	router    *gin.Engine
}

func (suite *TaskHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.mockRepo = &MockTaskRepository{}
	suite.mockCache = &MockCache{}
	suite.handler = handlers.NewTaskHandler(suite.mockRepo, suite.mockCache)
	suite.router = gin.New()

}

func TestTaskHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(TaskHandlerTestSuite))
}

func (suite *TaskHandlerTestSuite) TestCreateTask_Success() {
	// Setup
	reqBody := handlers.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      models.StatusPending,
		Assignee:    "test@example.com",
	}

	expectedID := uuid.New()
	suite.mockRepo.CreateFunc = func(task *models.Task) error {
		task.ID = expectedID
		task.CreatedAt = time.Now()
		task.UpdatedAt = time.Now()
		return nil
	}

	// Execute
	w := httptest.NewRecorder()
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.router.POST("/tasks", suite.handler.CreateTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response handlers.TaskResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedID, response.ID)
	assert.Equal(suite.T(), "Test Task", response.Title)
	assert.Equal(suite.T(), models.StatusPending, response.Status)
}

func (suite *TaskHandlerTestSuite) TestCreateTask_InvalidRequest() {
	// Setup - missing required title
	reqBody := handlers.CreateTaskRequest{
		Description: "Test Description",
	}

	// Execute
	w := httptest.NewRecorder()
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.router.POST("/tasks", suite.handler.CreateTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response.Error, "required")
}

func (suite *TaskHandlerTestSuite) TestGetTask_Success() {
	// Setup
	taskID := uuid.New()
	expectedTask := &models.Task{
		ID:          taskID,
		Title:       "Test Task",
		Description: "Test Description",
		Status:      models.StatusInProgress,
		Assignee:    "test@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	suite.mockRepo.GetByIDFunc = func(id uuid.UUID) (*models.Task, error) {
		if id == taskID {
			return expectedTask, nil
		}
		return nil, sql.ErrNoRows
	}

	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks/"+taskID.String(), nil)
	suite.router.GET("/tasks/:id", suite.handler.GetTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response handlers.TaskResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), taskID, response.ID)
	assert.Equal(suite.T(), "Test Task", response.Title)
	assert.Equal(suite.T(), models.StatusInProgress, response.Status)
}

func (suite *TaskHandlerTestSuite) TestGetTask_NotFound() {
	// Setup
	taskID := uuid.New()
	suite.mockRepo.GetByIDFunc = func(id uuid.UUID) (*models.Task, error) {
		return nil, sql.ErrNoRows
	}

	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks/"+taskID.String(), nil)
	suite.router.GET("/tasks/:id", suite.handler.GetTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Task not found", response.Error)
}

func (suite *TaskHandlerTestSuite) TestGetTask_InvalidID() {
	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks/invalid-id", nil)
	suite.router.GET("/tasks/:id", suite.handler.GetTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid task ID", response.Error)
}

func (suite *TaskHandlerTestSuite) TestGetTask_DatabaseError() {
	// Setup
	taskID := uuid.New()
	suite.mockRepo.GetByIDFunc = func(id uuid.UUID) (*models.Task, error) {
		return nil, assert.AnError
	}

	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks/"+taskID.String(), nil)
	suite.router.GET("/tasks/:id", suite.handler.GetTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Failed to get task", response.Error)
}

func (suite *TaskHandlerTestSuite) TestGetTasks_Success() {
	// Setup
	expectedTasks := []models.Task{
		{
			ID:          uuid.New(),
			Title:       "Task 1",
			Description: "Description 1",
			Status:      models.StatusPending,
			Assignee:    "user1@example.com",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Title:       "Task 2",
			Description: "Description 2",
			Status:      models.StatusCompleted,
			Assignee:    "user2@example.com",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	suite.mockRepo.GetAllFunc = func(page, limit int, status, assignee string) ([]models.Task, int64, error) {
		return expectedTasks, 2, nil
	}

	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks?page=1&limit=10", nil)
	suite.router.GET("/tasks", suite.handler.GetTasks)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response handlers.TaskListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(response.Tasks))
	assert.Equal(suite.T(), int64(2), response.Total)
	assert.Equal(suite.T(), 1, response.Page)
	assert.Equal(suite.T(), 10, response.Limit)
}

func (suite *TaskHandlerTestSuite) TestUpdateTask_Success() {
	// Setup
	taskID := uuid.New()
	existingTask := &models.Task{
		ID:          taskID,
		Title:       "Original Title",
		Description: "Original Description",
		Status:      models.StatusPending,
		Assignee:    "original@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	updateReq := handlers.UpdateTaskRequest{
		Title:  stringPtr("Updated Title"),
		Status: statusPtr(models.StatusCompleted),
	}

	suite.mockRepo.GetByIDFunc = func(id uuid.UUID) (*models.Task, error) {
		if id == taskID {
			return existingTask, nil
		}
		return nil, sql.ErrNoRows
	}

	suite.mockRepo.UpdateFunc = func(task *models.Task) error {
		// Simulate update
		return nil
	}

	// Execute
	w := httptest.NewRecorder()
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/tasks/"+taskID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.router.PUT("/tasks/:id", suite.handler.UpdateTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response handlers.TaskResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Title", response.Title)
	assert.Equal(suite.T(), models.StatusCompleted, response.Status)
}

func (suite *TaskHandlerTestSuite) TestDeleteTask_Success() {
	// Setup
	taskID := uuid.New()
	suite.mockRepo.DeleteFunc = func(id uuid.UUID) error {
		return nil
	}

	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/tasks/"+taskID.String(), nil)
	suite.router.DELETE("/tasks/:id", suite.handler.DeleteTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
}

func (suite *TaskHandlerTestSuite) TestDeleteTask_NotFound() {
	// Setup
	taskID := uuid.New()
	suite.mockRepo.DeleteFunc = func(id uuid.UUID) error {
		return sql.ErrNoRows
	}

	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/tasks/"+taskID.String(), nil)
	suite.router.DELETE("/tasks/:id", suite.handler.DeleteTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Task not found", response.Error)
}

func (suite *TaskHandlerTestSuite) TestDeleteTask_InvalidID() {
	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/tasks/invalid-id", nil)
	suite.router.DELETE("/tasks/:id", suite.handler.DeleteTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid task ID", response.Error)
}

func (suite *TaskHandlerTestSuite) TestDeleteTask_DatabaseError() {
	// Setup
	taskID := uuid.New()
	suite.mockRepo.DeleteFunc = func(id uuid.UUID) error {
		return assert.AnError
	}

	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/tasks/"+taskID.String(), nil)
	suite.router.DELETE("/tasks/:id", suite.handler.DeleteTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Failed to delete task", response.Error)
}

func (suite *TaskHandlerTestSuite) TestUpdateTask_NotFound() {
	// Setup
	taskID := uuid.New()
	updateReq := handlers.UpdateTaskRequest{
		Title: stringPtr("Updated Title"),
	}

	suite.mockRepo.GetByIDFunc = func(id uuid.UUID) (*models.Task, error) {
		return nil, sql.ErrNoRows
	}

	// Execute
	w := httptest.NewRecorder()
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/tasks/"+taskID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.router.PUT("/tasks/:id", suite.handler.UpdateTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Task not found", response.Error)
}

func (suite *TaskHandlerTestSuite) TestUpdateTask_InvalidID() {
	// Setup
	updateReq := handlers.UpdateTaskRequest{
		Title: stringPtr("Updated Title"),
	}

	// Execute
	w := httptest.NewRecorder()
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/tasks/invalid-id", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.router.PUT("/tasks/:id", suite.handler.UpdateTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response handlers.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Invalid task ID", response.Error)
}

func (suite *TaskHandlerTestSuite) TestUpdateTask_InvalidRequest() {
	// Setup
	taskID := uuid.New()
	// Invalid JSON
	invalidBody := []byte(`{"title":}`)

	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/tasks/"+taskID.String(), bytes.NewBuffer(invalidBody))
	req.Header.Set("Content-Type", "application/json")
	suite.router.PUT("/tasks/:id", suite.handler.UpdateTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *TaskHandlerTestSuite) TestUpdateTask_DatabaseError() {
	// Setup
	taskID := uuid.New()
	existingTask := &models.Task{
		ID:          taskID,
		Title:       "Original Title",
		Description: "Original Description",
		Status:      models.StatusPending,
		Assignee:    "original@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	updateReq := handlers.UpdateTaskRequest{
		Title: stringPtr("Updated Title"),
	}

	suite.mockRepo.GetByIDFunc = func(id uuid.UUID) (*models.Task, error) {
		if id == taskID {
			return existingTask, nil
		}
		return nil, sql.ErrNoRows
	}

	suite.mockRepo.UpdateFunc = func(task *models.Task) error {
		return assert.AnError
	}

	// Execute
	w := httptest.NewRecorder()
	body, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/tasks/"+taskID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.router.PUT("/tasks/:id", suite.handler.UpdateTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
}

func (suite *TaskHandlerTestSuite) TestGetTasks_InvalidPage() {
	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks?page=invalid", nil)
	suite.router.GET("/tasks", suite.handler.GetTasks)
	suite.router.ServeHTTP(w, req)

	// Assert - should default to page 1
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *TaskHandlerTestSuite) TestGetTasks_InvalidLimit() {
	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks?limit=invalid", nil)
	suite.router.GET("/tasks", suite.handler.GetTasks)
	suite.router.ServeHTTP(w, req)

	// Assert - should default to limit 10
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *TaskHandlerTestSuite) TestCreateTask_DatabaseError() {
	// Setup
	reqBody := handlers.CreateTaskRequest{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      models.StatusPending,
		Assignee:    "test@example.com",
	}

	suite.mockRepo.CreateFunc = func(task *models.Task) error {
		return assert.AnError
	}

	// Execute
	w := httptest.NewRecorder()
	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.router.POST("/tasks", suite.handler.CreateTask)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
}

func (suite *TaskHandlerTestSuite) TestGetTasks_WithFilters() {
	// Setup
	expectedTasks := []models.Task{
		{
			ID:          uuid.New(),
			Title:       "Task 1",
			Description: "Description 1",
			Status:      models.StatusPending,
			Assignee:    "user1@example.com",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	suite.mockRepo.GetAllFunc = func(page, limit int, status, assignee string) ([]models.Task, int64, error) {
		assert.Equal(suite.T(), 1, page)
		assert.Equal(suite.T(), 5, limit)
		assert.Equal(suite.T(), "pending", status)
		assert.Equal(suite.T(), "user1@example.com", assignee)
		return expectedTasks, 1, nil
	}

	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks?page=1&limit=5&status=pending&assignee=user1@example.com", nil)
	suite.router.GET("/tasks", suite.handler.GetTasks)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response handlers.TaskListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(response.Tasks))
	assert.Equal(suite.T(), int64(1), response.Total)
}

func (suite *TaskHandlerTestSuite) TestGetTasks_DatabaseError() {
	// Setup
	suite.mockRepo.GetAllFunc = func(page, limit int, status, assignee string) ([]models.Task, int64, error) {
		return nil, 0, assert.AnError
	}

	// Execute
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks?page=1&limit=10", nil)
	suite.router.GET("/tasks", suite.handler.GetTasks)
	suite.router.ServeHTTP(w, req)

	// Assert
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)
}
