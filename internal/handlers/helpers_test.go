package handlers

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"taheri24.ir/graph1/internal/models"
	"taheri24.ir/graph1/internal/types"
)

func TestTasksToResponses(t *testing.T) {
	// Create test tasks with known timestamps
	baseTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	tasks := []models.Task{
		{
			ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Title:       "Test Task 1",
			Description: "Test Description 1",
			Status:      types.StatusPending,
			Assignee:    "Test Assignee 1",
			CreatedAt:   baseTime,
			UpdatedAt:   baseTime.Add(time.Hour),
		},
		{
			ID:          uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			Title:       "Test Task 2",
			Description: "Test Description 2",
			Status:      types.StatusInProgress,
			Assignee:    "Test Assignee 2",
			CreatedAt:   baseTime.Add(2 * time.Hour),
			UpdatedAt:   baseTime.Add(3 * time.Hour),
		},
	}

	responses := tasksToResponses(tasks)

	// Verify conversion
	assert.Len(t, responses, 2)

	// Check first task
	assert.Equal(t, tasks[0].ID, responses[0].ID)
	assert.Equal(t, "Test Task 1", responses[0].Title)
	assert.Equal(t, "Test Description 1", responses[0].Description)
	assert.Equal(t, types.StatusPending, responses[0].Status)
	assert.Equal(t, "Test Assignee 1", responses[0].Assignee)
	assert.Equal(t, "2023-01-01T12:00:00Z", responses[0].CreatedAt)
	assert.Equal(t, "2023-01-01T13:00:00Z", responses[0].UpdatedAt)

	// Check second task
	assert.Equal(t, tasks[1].ID, responses[1].ID)
	assert.Equal(t, "Test Task 2", responses[1].Title)
	assert.Equal(t, "Test Description 2", responses[1].Description)
	assert.Equal(t, types.StatusInProgress, responses[1].Status)
	assert.Equal(t, "Test Assignee 2", responses[1].Assignee)
	assert.Equal(t, "2023-01-01T14:00:00Z", responses[1].CreatedAt)
	assert.Equal(t, "2023-01-01T15:00:00Z", responses[1].UpdatedAt)
}

func TestTasksToResponses_EmptySlice(t *testing.T) {
	tasks := []models.Task{}
	responses := tasksToResponses(tasks)

	assert.Len(t, responses, 0)
}

func TestFilterTasksByStatus(t *testing.T) {
	tasks := []models.Task{
		{
			ID:       uuid.New(),
			Title:    "Pending Task",
			Status:   types.StatusPending,
			Assignee: "User1",
		},
		{
			ID:       uuid.New(),
			Title:    "In Progress Task",
			Status:   types.StatusInProgress,
			Assignee: "User2",
		},
		{
			ID:       uuid.New(),
			Title:    "Completed Task",
			Status:   types.StatusCompleted,
			Assignee: "User3",
		},
		{
			ID:       uuid.New(),
			Title:    "Another Pending Task",
			Status:   types.StatusPending,
			Assignee: "User4",
		},
	}

	// Test filtering by pending status
	pendingTasks := filterTasksByStatus(tasks, types.StatusPending)
	assert.Len(t, pendingTasks, 2)
	assert.Equal(t, "Pending Task", pendingTasks[0].Title)
	assert.Equal(t, "Another Pending Task", pendingTasks[1].Title)

	// Test filtering by in progress status
	inProgressTasks := filterTasksByStatus(tasks, types.StatusInProgress)
	assert.Len(t, inProgressTasks, 1)
	assert.Equal(t, "In Progress Task", inProgressTasks[0].Title)

	// Test filtering by completed status
	completedTasks := filterTasksByStatus(tasks, types.StatusCompleted)
	assert.Len(t, completedTasks, 1)
	assert.Equal(t, "Completed Task", completedTasks[0].Title)

	// Test filtering by status that doesn't exist
	cancelledTasks := filterTasksByStatus(tasks, types.TaskStatus("cancelled"))
	assert.Len(t, cancelledTasks, 0)
}

func TestFilterTasksByStatus_EmptySlice(t *testing.T) {
	tasks := []models.Task{}
	filtered := filterTasksByStatus(tasks, types.StatusPending)

	assert.Len(t, filtered, 0)
}

func TestFilterTasksByAssignee(t *testing.T) {
	tasks := []models.Task{
		{
			ID:       uuid.New(),
			Title:    "Task 1",
			Status:   types.StatusPending,
			Assignee: "Alice",
		},
		{
			ID:       uuid.New(),
			Title:    "Task 2",
			Status:   types.StatusInProgress,
			Assignee: "Bob",
		},
		{
			ID:       uuid.New(),
			Title:    "Task 3",
			Status:   types.StatusCompleted,
			Assignee: "Alice",
		},
		{
			ID:       uuid.New(),
			Title:    "Task 4",
			Status:   types.StatusPending,
			Assignee: "Charlie",
		},
	}

	// Test filtering by Alice
	aliceTasks := filterTasksByAssignee(tasks, "Alice")
	assert.Len(t, aliceTasks, 2)
	assert.Equal(t, "Task 1", aliceTasks[0].Title)
	assert.Equal(t, "Task 3", aliceTasks[1].Title)

	// Test filtering by Bob
	bobTasks := filterTasksByAssignee(tasks, "Bob")
	assert.Len(t, bobTasks, 1)
	assert.Equal(t, "Task 2", bobTasks[0].Title)

	// Test filtering by Charlie
	charlieTasks := filterTasksByAssignee(tasks, "Charlie")
	assert.Len(t, charlieTasks, 1)
	assert.Equal(t, "Task 4", charlieTasks[0].Title)

	// Test filtering by assignee that doesn't exist
	daveTasks := filterTasksByAssignee(tasks, "Dave")
	assert.Len(t, daveTasks, 0)
}

func TestFilterTasksByAssignee_EmptySlice(t *testing.T) {
	tasks := []models.Task{}
	filtered := filterTasksByAssignee(tasks, "Anyone")

	assert.Len(t, filtered, 0)
}

func TestFilterTasksByAssignee_EmptyAssignee(t *testing.T) {
	tasks := []models.Task{
		{
			ID:       uuid.New(),
			Title:    "Task 1",
			Status:   types.StatusPending,
			Assignee: "Alice",
		},
		{
			ID:       uuid.New(),
			Title:    "Task 2",
			Status:   types.StatusInProgress,
			Assignee: "",
		},
		{
			ID:       uuid.New(),
			Title:    "Task 3",
			Status:   types.StatusCompleted,
			Assignee: "",
		},
	}

	// Test filtering by empty string
	emptyAssigneeTasks := filterTasksByAssignee(tasks, "")
	assert.Len(t, emptyAssigneeTasks, 2)
	assert.Equal(t, "Task 2", emptyAssigneeTasks[0].Title)
	assert.Equal(t, "Task 3", emptyAssigneeTasks[1].Title)
}
