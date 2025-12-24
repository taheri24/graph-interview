package handlers

import (
	"net/http"
	"strconv"

	"taheri24.ir/graph1/internal/cache"
	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/dto"
	"taheri24.ir/graph1/internal/middleware"
	"taheri24.ir/graph1/internal/models"
	"taheri24.ir/graph1/internal/types"
	"taheri24.ir/graph1/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskHandler struct {
	repo  database.TaskRepository
	cache cache.CacheInterface[models.Task]
}

func NewTaskHandler(repo database.TaskRepository, cache cache.CacheInterface[models.Task]) *TaskHandler {
	return &TaskHandler{repo: repo, cache: cache}
}

// CreateTask handles POST /tasks
// @Summary Create a new task
// @Description Create a new task with the provided information
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body dto.CreateTaskRequest true "Task information"
// @Success 201 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger := middleware.GetLoggerFromContext(c.Request.Context())
		logger.Error("Invalid request body for creating task", "error", err)
		c.JSON(http.StatusBadRequest, dto.NewErr(err))
		return
	}

	task := models.Task{
		ID:          uuid.New(),
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Assignee:    req.Assignee,
	}

	if req.Status == "" {
		task.Status = types.StatusPending
	}

	if err := h.repo.Create(c.Request.Context(), &task); err != nil {
		logger := middleware.GetLoggerFromContext(c.Request.Context())
		logger.Error("Failed to create task in repository", "title", req.Title, "error", err)
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to create task"))
		return
	}

	response := dto.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Assignee:    task.Assignee,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	logger := middleware.GetLoggerFromContext(c.Request.Context())
	logger.Info("Task created successfully", "id", task.ID.String(), "title", task.Title, "status", string(task.Status))

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
// @Success 200 {object} dto.TaskListResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks [get]
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

	tasks, total, err := h.repo.GetAll(c.Request.Context(), page, limit, status, assignee)
	if err != nil {
		logger := middleware.GetLoggerFromContext(c.Request.Context())
		logger.Error("Failed to fetch tasks from repository", "page", page, "limit", limit, "status", status, "assignee", assignee, "error", err)
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to fetch tasks"))
		return
	}

	taskResponses := make([]dto.TaskResponse, len(tasks))
	for i, task := range tasks {
		taskResponses[i] = dto.TaskResponse{
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

	response := dto.TaskListResponse{
		Tasks:       taskResponses,
		Total:       total,
		Page:        page,
		Limit:       limit,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}

	logger := middleware.GetLoggerFromContext(c.Request.Context())
	logger.Info("Tasks retrieved successfully", "page", page, "limit", limit, "total", total, "status", status, "assignee", assignee)

	c.JSON(http.StatusOK, response)
}

// GetTask handles GET /tasks/{id}
// @Summary Get a task by ID
// @Description Retrieve a specific task by its UUID
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID (UUID)"
// @Success 200 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/v1/tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger := middleware.GetLoggerFromContext(c.Request.Context())
		logger.Error("Invalid task ID provided", "idStr", idStr, "error", err)
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid task ID"))
		return
	}

	// Try to get from cache first
	taskPtr, err := h.cache.Get(id.String())
	if err == nil && taskPtr != nil {
		// Cache hit
		logger := middleware.GetLoggerFromContext(c.Request.Context())
		logger.Info("Task retrieved from cache", "id", id.String())
		c.Header("X-Cache-Status", "HIT")
		response := dto.TaskResponse{
			ID:          taskPtr.ID,
			Title:       taskPtr.Title,
			Description: taskPtr.Description,
			Status:      taskPtr.Status,
			Assignee:    taskPtr.Assignee,
			CreatedAt:   taskPtr.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   taskPtr.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// Cache miss, get from repository
	taskPtr, err = h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		logger := middleware.GetLoggerFromContext(c.Request.Context())
		if utils.ErrIsRecordNotFound(err) {
			logger.Info("Task not found", "id", id.String())
			c.Header("X-Cache-Status", "MISS")
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Task not found"))
		} else {
			logger.Error("Failed to get task from repository", "id", id.String(), "error", err)
			c.Header("X-Cache-Status", "MISS")
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to get task"))
		}
		return
	}

	// Set in cache
	if err := h.cache.Set(id.String(), *taskPtr); err != nil {
		// Log error but don't fail the request
		logger := middleware.GetLoggerFromContext(c.Request.Context())
		logger.Error("Failed to set task in cache", "id", id.String(), "error", err)
	}

	logger := middleware.GetLoggerFromContext(c.Request.Context())
	logger.Info("Task retrieved from database", "id", id.String())

	c.Header("X-Cache-Status", "MISS")
	response := dto.TaskResponse{
		ID:          taskPtr.ID,
		Title:       taskPtr.Title,
		Description: taskPtr.Description,
		Status:      taskPtr.Status,
		Assignee:    taskPtr.Assignee,
		CreatedAt:   taskPtr.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   taskPtr.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
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
// @Param task body dto.UpdateTaskRequest true "Task update information"
// @Success 200 {object} dto.TaskResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger := middleware.GetLoggerFromContext(c.Request.Context())
		logger.Error("Invalid task ID provided", "idStr", idStr, "error", err)
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid task ID"))
		return
	}

	task, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		logger := middleware.GetLoggerFromContext(c.Request.Context())
		if utils.ErrIsRecordNotFound(err) {
			logger.Info("Task not found for update", "id", id.String())
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Task not found"))
		} else {
			logger.Error("Failed to get task for update", "id", id.String(), "error", err)
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to get task"))
		}
		return
	}

	var req dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger := middleware.GetLoggerFromContext(c.Request.Context())
		logger.Error("Invalid request body for updating task", "id", id.String(), "error", err)
		c.JSON(http.StatusBadRequest, dto.NewErr(err))
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

	if err := h.repo.Update(c.Request.Context(), task); err != nil {
		logger := middleware.GetLoggerFromContext(c.Request.Context())
		logger.Error("Failed to update task in repository", "id", id.String(), "error", err)
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to update task"))
		return
	}

	// Invalidate cache
	if err := h.cache.Invalidate(id.String()); err != nil {
		// Log error but don't fail the request
		logger := middleware.GetLoggerFromContext(c.Request.Context())
		logger.Error("Failed to invalidate task cache", "id", id.String(), "error", err)
	}

	response := dto.TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Assignee:    task.Assignee,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	logger := middleware.GetLoggerFromContext(c.Request.Context())
	logger.Info("Task updated successfully", "id", task.ID.String(), "title", task.Title, "status", string(task.Status))

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
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger := middleware.GetLoggerFromContext(c.Request.Context())
		logger.Error("Invalid task ID provided", "idStr", idStr, "error", err)
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid task ID"))
		return
	}
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		logger := middleware.GetLoggerFromContext(c.Request.Context())
		if utils.ErrIsRecordNotFound(err) {
			logger.Info("Task not found for deletion", "id", id.String())
			c.JSON(http.StatusNotFound, dto.NewErrorResponse("Task not found"))
		} else {
			logger.Error("Failed to delete task from repository", "id", id.String(), "error", err)
			c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("Failed to delete task"))
		}
		return
	}

	// Invalidate cache
	if err := h.cache.Invalidate(id.String()); err != nil {
		// Log error but don't fail the request
		logger := middleware.GetLoggerFromContext(c.Request.Context())
		logger.Error("Failed to invalidate task cache", "id", id.String(), "error", err)
	}

	logger := middleware.GetLoggerFromContext(c.Request.Context())
	logger.Info("Task deleted successfully", "id", id.String())

	c.JSON(http.StatusNoContent, nil)
}
