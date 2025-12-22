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
	db *database.Database
}

func NewTaskHandler(db *database.Database) *TaskHandler {
	return &TaskHandler{db: db}
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

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// CreateTask handles POST /tasks
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
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

	if err := h.db.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create task"})
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

	offset := (page - 1) * limit

	query := h.db.DB.Model(&models.Task{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if assignee != "" {
		query = query.Where("assignee = ?", assignee)
	}

	var total int64
	query.Count(&total)

	var tasks []models.Task
	if err := query.Offset(offset).Limit(limit).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch tasks"})
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
func (h *TaskHandler) GetTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid task ID"})
		return
	}

	var task models.Task
	if err := h.db.DB.First(&task, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Task not found"})
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
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid task ID"})
		return
	}

	var task models.Task
	if err := h.db.DB.First(&task, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Task not found"})
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
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

	if err := h.db.DB.Save(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update task"})
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
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid task ID"})
		return
	}

	result := h.db.DB.Delete(&models.Task{}, "id = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete task"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "Task not found"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
