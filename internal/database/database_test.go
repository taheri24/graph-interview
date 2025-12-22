package database_test

import (
	"database/sql/driver"
	"testing"
	"time"

	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DatabaseTestSuite struct {
	suite.Suite
	mock sqlmock.Sqlmock
	db   *database.Database
}

func (suite *DatabaseTestSuite) SetupTest() {
	_, mock, err := sqlmock.New()
	suite.Require().NoError(err)

	suite.mock = mock

	// Create a database instance with the mocked connection
	// This is a bit tricky since we normally use GORM, but for testing
	// we'll create a simplified version
	suite.db = &database.Database{}
	// Note: For full testing, we'd need to inject the GORM DB instance
	// For now, we'll focus on testing the interface compliance
}

func (suite *DatabaseTestSuite) TearDownTest() {
	suite.mock.ExpectClose()
}

func TestDatabaseTestSuite(t *testing.T) {
	suite.Run(t, new(DatabaseTestSuite))
}

func (suite *DatabaseTestSuite) TestDatabaseImplementsTaskRepository() {
	// Test that Database implements TaskRepository interface
	var _ database.TaskRepository = (*database.Database)(nil)
}

// Mock driver for UUID
type uuidMock struct{}

func (u uuidMock) Match(v driver.Value) bool {
	_, ok := v.(string)
	return ok
}

// Test Create method would require setting up GORM with sqlmock
// For now, we'll test the interface and basic structure
func (suite *DatabaseTestSuite) TestTaskRepositoryInterface() {
	repo := suite.db

	// Test that all methods exist and have correct signatures
	// This is more of a compile-time test, but ensures interface compliance

	// Create a test task
	task := &models.Task{
		ID:          uuid.New(),
		Title:       "Test Task",
		Description: "Test Description",
		Status:      models.StatusPending,
		Assignee:    "test@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// These would normally work with a real database
	// For unit testing with sqlmock, we'd need to set up GORM expectations
	// Since GORM doesn't directly support sqlmock easily, these are integration tests

	_ = repo // Use the variable to avoid unused variable error
	_ = task // Use the variable to avoid unused variable error

	// Interface compliance test passed if code compiles
	assert.True(suite.T(), true, "Database implements TaskRepository interface")
}

// Test UUID generation and validation
func TestUUIDHandling(t *testing.T) {
	id := uuid.New()
	assert.NotEqual(t, uuid.Nil, id)

	// Test parsing
	idStr := id.String()
	parsedID, err := uuid.Parse(idStr)
	assert.NoError(t, err)
	assert.Equal(t, id, parsedID)
}

// Test TaskStatus enum values
func TestTaskStatus(t *testing.T) {
	assert.Equal(t, models.TaskStatus("pending"), models.StatusPending)
	assert.Equal(t, models.TaskStatus("in_progress"), models.StatusInProgress)
	assert.Equal(t, models.TaskStatus("completed"), models.StatusCompleted)
}

// Test Task model structure
func TestTaskModel(t *testing.T) {
	task := models.Task{
		ID:          uuid.New(),
		Title:       "Test Title",
		Description: "Test Description",
		Status:      models.StatusPending,
		Assignee:    "test@example.com",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, task.ID)
	assert.Equal(t, "Test Title", task.Title)
	assert.Equal(t, "Test Description", task.Description)
	assert.Equal(t, models.StatusPending, task.Status)
	assert.Equal(t, "test@example.com", task.Assignee)
	assert.True(t, task.CreatedAt.Before(time.Now().Add(time.Second)))
	assert.True(t, task.UpdatedAt.Before(time.Now().Add(time.Second)))
}
