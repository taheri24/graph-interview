package handlers

import (
	"net/http"
	"strconv"

	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskHandler struct {
	repo database.TaskRepository
}

func NewTaskHandler(repo database.TaskRepository) *TaskHandler {
	return &TaskHandler{repo: repo}
}

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Title       string            `json:"title" binding:"required,min=1,max=200"`
	Description string            `json:"description" binding:"max=1000"`
	Status      models.TaskStatus `json:"status" binding:"omitempty,oneof=pending in_progress completed"`
	Assignee    string            `json:"assignee" binding:"max=100"`
}

// UpdateTaskRequest represents the request body for updating a task
type UpdateTaskRequest struct {
	Title       *string            `json:"title" binding:"omitempty,min=1,max=200"`
	Description *string            `json:"description" binding:"omitempty,max=1000"`
	Status      *models.TaskStatus `json:"status" binding:"omitempty,oneof=pending in_progress completed"`
	Assignee    *string            `json:"assignee" binding:"omitempty,max=100"`
}

// TaskResponse represents the response body for a task
type TaskResponse struct {
	ID          uuid.UUID         `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      models.TaskStatus `json:"status"`
	Assignee    string            `json:"assignee"`
	CreatedAt   string            `json:"created_at"`
	UpdatedAt   string            `json:"updated_at"`
}

// TaskListResponse represents the response body for listing tasks
type TaskListResponse struct {
	Tasks       []TaskResponse `json:"tasks"`
	Total       int64          `json:"total"`
	Page        int            `json:"page"`
	Limit       int            `json:"limit"`
	HasNext     bool           `json:"has_next"`
	HasPrevious bool           `json:"has_previous"`
}

// CreateTask handles POST /tasks
// @Summary Create a new task
// @Description Create a new task with the provided information
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body CreateTaskRequest true "Task information"
// @Success 201 {object} TaskResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErr(err))
		return
	}

	task := models.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Assignee:    req.Assignee,
	}

	if req.Status == "" {
		task.Status = models.StatusPending
	}

	if err := h.repo.Create(&task); err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to create task"))
		return
	}

	response := TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Assignee:    task.Assignee,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusCreated, response)
}

// GetTasks handles GET /tasks
// @Summary Get all tasks with pagination and filtering
// @Description Retrieve a paginated list of tasks with optional filtering by status and assignee
// @Tags tasks
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)" minimum(1)
// @Param limit query int false "Items per page (default: 10, max: 100)" minimum(1) maximum(100)
// @Param status query string false "Filter by status (pending, in_progress, completed)"
// @Param assignee query string false "Filter by assignee"
// @Success 200 {object} TaskListResponse
// @Failure 500 {object} ErrorResponse
// @Router /tasks [get]
func (h *TaskHandler) GetTasks(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	status := c.Query("status")
	assignee := c.Query("assignee")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	tasks, total, err := h.repo.GetAll(page, limit, status, assignee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to fetch tasks"))
		return
	}

	taskResponses := make([]TaskResponse, len(tasks))
	for i, task := range tasks {
		taskResponses[i] = TaskResponse{
			ID:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			Status:      task.Status,
			Assignee:    task.Assignee,
			CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	response := TaskListResponse{
		Tasks:       taskResponses,
		Total:       total,
		Page:        page,
		Limit:       limit,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}

	c.JSON(http.StatusOK, response)
}

// GetTask handles GET /tasks/{id}
// @Summary Get a task by ID
// @Description Retrieve a specific task by its UUID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID (UUID)"
// @Success 200 {object} TaskResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid task ID"))
		return
	}

	task, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, NewErrorResponse("Task not found"))
		return
	}

	response := TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Assignee:    task.Assignee,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, response)
}

// UpdateTask handles PUT /tasks/{id}
// @Summary Update a task
// @Description Update an existing task with the provided information. Only provided fields will be updated.
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID (UUID)"
// @Param task body UpdateTaskRequest true "Task update information"
// @Success 200 {object} TaskResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid task ID"))
		return
	}

	task, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, NewErrorResponse("Task not found"))
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErr(err))
		return
	}

	// Update only provided fields
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	if req.Assignee != nil {
		task.Assignee = *req.Assignee
	}

	if err := h.repo.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to update task"))
		return
	}

	response := TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Assignee:    task.Assignee,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, response)
}

// DeleteTask handles DELETE /tasks/{id}
// @Summary Delete a task
// @Description Delete a task by its UUID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid task ID"))
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to delete task"))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
