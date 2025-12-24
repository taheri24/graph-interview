package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"taheri24.ir/graph1/internal/cache"
	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/dto"
	"taheri24.ir/graph1/internal/handlers"
	"taheri24.ir/graph1/internal/models"
	"taheri24.ir/graph1/internal/types"
	"taheri24.ir/graph1/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite runs tests against a real database
type IntegrationTestSuite struct {
	suite.Suite
	db     *database.Database
	router *gin.Engine
}

func (suite *IntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	// Use in-memory test database configuration
	testConfig := config.NewTestConfig()

	var err error
	suite.db, err = database.NewDatabase(testConfig)
	if err != nil {
		suite.T().Skipf("Skipping integration tests: %v", err)
		return
	}

	// Clean up any existing data
	suite.db.DB.Exec("DELETE FROM tasks")

	// Setup cache
	redisAddr := fmt.Sprintf("%s:%s", testConfig.Redis.Host, testConfig.Redis.Port)
	redisCache, err := cache.NewRedisCache(redisAddr, testConfig.Redis.Password, testConfig.Redis.DB)
	if err != nil {
		suite.T().Skipf("Skipping integration tests: Redis not available: %v", err)
		return
	}
	taskCache := cache.NewRedisCacheImpl[models.Task]("tasks", redisCache)

	// Setup router
	suite.router = gin.New()
	taskHandler := handlers.NewTaskHandler(suite.db, taskCache)

	api := suite.router.Group("/tasks")
	{
		api.POST("", taskHandler.CreateTask)
		api.GET("", taskHandler.GetTasks)
		api.GET("/:id", taskHandler.GetTask)
		api.PUT("/:id", taskHandler.UpdateTask)
		api.DELETE("/:id", taskHandler.DeleteTask)
	}
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.db.DB.Exec("DELETE FROM tasks")
		suite.db.Close()
	}
}

func (suite *IntegrationTestSuite) TearDownTest() {
	if suite.db != nil {
		suite.db.DB.Exec("DELETE FROM tasks")
	}
}

func TestIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(IntegrationTestSuite))
}

func (suite *IntegrationTestSuite) TestCreateAndGetTask() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}

	// Create a task
	createReq := dto.CreateTaskRequest{
		Title:       "Integration Test Task",
		Description: "Testing integration",
		Status:      types.StatusPending,
		Assignee:    "integration@example.com",
	}

	w := httptest.NewRecorder()
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createResp dto.TaskResponse
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	assert.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), uuid.Nil, createResp.ID)
	assert.Equal(suite.T(), "Integration Test Task", createResp.Title)
	assert.Equal(suite.T(), types.StatusPending, createResp.Status)

	// Get the task back
	taskID := createResp.ID.String()
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/tasks/"+taskID, nil)
	suite.router.ServeHTTP(w2, req2)

	assert.Equal(suite.T(), http.StatusOK, w2.Code)

	var getResp dto.TaskResponse
	err = json.Unmarshal(w2.Body.Bytes(), &getResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), createResp.ID, getResp.ID)
	assert.Equal(suite.T(), "Integration Test Task", getResp.Title)
}

func (suite *IntegrationTestSuite) TestUpdateTask() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}

	// First create a task
	createReq := dto.CreateTaskRequest{
		Title:       "Original Title",
		Description: "Original Description",
		Status:      types.StatusPending,
		Assignee:    "original@example.com",
	}

	w := httptest.NewRecorder()
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	var createResp dto.TaskResponse
	json.Unmarshal(w.Body.Bytes(), &createResp)
	taskID := createResp.ID.String()

	// Update the task
	updateReq := dto.UpdateTaskRequest{
		Title:  stringPtr("Updated Title"),
		Status: statusPtr(types.StatusCompleted),
	}

	w2 := httptest.NewRecorder()
	body2, _ := json.Marshal(updateReq)
	req2, _ := http.NewRequest("PUT", "/tasks/"+taskID, bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w2, req2)

	assert.Equal(suite.T(), http.StatusOK, w2.Code)

	var updateResp dto.TaskResponse
	err := json.Unmarshal(w2.Body.Bytes(), &updateResp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Title", updateResp.Title)
	assert.Equal(suite.T(), types.StatusCompleted, updateResp.Status)
	assert.Equal(suite.T(), "original@example.com", updateResp.Assignee) // Should remain unchanged
}

func (suite *IntegrationTestSuite) TestDeleteTask() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}

	// Create a task
	createReq := dto.CreateTaskRequest{
		Title:  "Task to Delete",
		Status: types.StatusPending,
	}

	w := httptest.NewRecorder()
	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	suite.router.ServeHTTP(w, req)

	var createResp dto.TaskResponse
	json.Unmarshal(w.Body.Bytes(), &createResp)
	taskID := createResp.ID.String()

	// Delete the task
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("DELETE", "/tasks/"+taskID, nil)
	suite.router.ServeHTTP(w2, req2)

	assert.Equal(suite.T(), http.StatusNoContent, w2.Code)

	// Verify it's gone
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/tasks/"+taskID, nil)
	suite.router.ServeHTTP(w3, req3)

	assert.Equal(suite.T(), http.StatusNotFound, w3.Code)
}

func (suite *IntegrationTestSuite) TestGetTasksWithPagination() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}

	// Create multiple tasks
	tasks := []dto.CreateTaskRequest{
		{Title: "Task 1", Status: types.StatusPending},
		{Title: "Task 2", Status: types.StatusInProgress},
		{Title: "Task 3", Status: types.StatusCompleted},
	}

	for _, task := range tasks {
		body, _ := json.Marshal(task)
		req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusCreated, w.Code)
	}

	// Get tasks with pagination
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks?page=1&limit=2", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var resp dto.TaskListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(resp.Tasks))
	assert.Equal(suite.T(), int64(3), resp.Total)
	assert.Equal(suite.T(), 1, resp.Page)
	assert.Equal(suite.T(), 2, resp.Limit)
}

func (suite *IntegrationTestSuite) TestGetTasksWithFiltering() {
	if suite.db == nil {
		suite.T().Skip("Database not available")
	}

	// Create tasks with different statuses
	tasks := []dto.CreateTaskRequest{
		{Title: "Pending Task", Status: types.StatusPending, Assignee: "user1@example.com"},
		{Title: "In Progress Task", Status: types.StatusInProgress, Assignee: "user2@example.com"},
		{Title: "Completed Task", Status: types.StatusCompleted, Assignee: "user1@example.com"},
	}

	for _, task := range tasks {
		body, _ := json.Marshal(task)
		req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		suite.router.ServeHTTP(w, req)
		assert.Equal(suite.T(), http.StatusCreated, w.Code)
	}

	// Filter by status
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/tasks?status=pending", nil)
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var resp dto.TaskListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(resp.Tasks))
	assert.Equal(suite.T(), "Pending Task", resp.Tasks[0].Title)

	// Filter by assignee
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/tasks?assignee=user1@example.com", nil)
	suite.router.ServeHTTP(w2, req2)

	assert.Equal(suite.T(), http.StatusOK, w2.Code)

	var resp2 dto.TaskListResponse
	err = json.Unmarshal(w2.Body.Bytes(), &resp2)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(resp2.Tasks))
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func statusPtr(s types.TaskStatus) *types.TaskStatus {
	return &s
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
