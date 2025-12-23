package handlers

import (
	"net/http"
	"strconv"
	"time"

	"taheri24.ir/graph1/internal/cache"
	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/middleware"
	"taheri24.ir/graph1/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CachedTaskHandler struct {
	repo  database.TaskRepository
	cache *cache.RedisCache
}

func NewCachedTaskHandler(repo database.TaskRepository, cache *cache.RedisCache) *CachedTaskHandler {
	return &CachedTaskHandler{
		repo:  repo,
		cache: cache,
	}
}

// GetTasks handles GET /tasks with Redis caching
func (h *CachedTaskHandler) GetTasks(c *gin.Context) {
	// Try to get from cache first
	tasks, err := h.cache.GetTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Cache error"))
		return
	}

	// Cache miss - fetch from database
	if tasks == nil {
		// Get pagination parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		status := c.Query("status")
		assignee := c.Query("assignee")

		if page < 1 {
			page = 1
		}
		if limit < 1 || limit > 100 {
			limit = 10
		}

		tasks, total, err := h.repo.GetAll(page, limit, status, assignee)
		if err != nil {
			c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to fetch tasks"))
			return
		}

		// Cache the result for 5 minutes (only cache the tasks, not pagination metadata)
		if err := h.cache.SetTasks(tasks, 5*time.Minute); err != nil {
			// Log error but don't fail the request
		}

		// Update tasks count metric
		middleware.UpdateTasksCount(float64(total))

		response := TaskListResponse{
			Tasks:       tasksToResponses(tasks),
			Total:       total,
			Page:        page,
			Limit:       limit,
			HasNext:     int64(page*limit) < total,
			HasPrevious: page > 1,
		}

		c.JSON(http.StatusOK, response)
		return
	}

	// For cached results, we need to apply filtering and pagination manually
	status := c.Query("status")
	assignee := c.Query("assignee")

	filteredTasks := tasks
	if status != "" {
		filteredTasks = filterTasksByStatus(filteredTasks, models.TaskStatus(status))
	}
	if assignee != "" {
		filteredTasks = filterTasksByAssignee(filteredTasks, assignee)
	}

	// Pagination for cached results
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	start := (page - 1) * limit
	end := start + limit
	if end > len(filteredTasks) {
		end = len(filteredTasks)
	}
	if start >= len(filteredTasks) {
		filteredTasks = []models.Task{}
	} else {
		filteredTasks = filteredTasks[start:end]
	}

	// Update tasks count metric
	middleware.UpdateTasksCount(float64(len(tasks)))

	response := TaskListResponse{
		Tasks:       tasksToResponses(filteredTasks),
		Total:       int64(len(tasks)),
		Page:        page,
		Limit:       limit,
		HasNext:     int64(page*limit) < int64(len(tasks)),
		HasPrevious: page > 1,
	}

	c.JSON(http.StatusOK, response)
}

// GetTask handles GET /tasks/{id} with Redis caching
func (h *CachedTaskHandler) GetTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid task ID"))
		return
	}

	// Try to get from cache first
	task, err := h.cache.GetTask(id.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Cache error"))
		return
	}

	// Cache miss - fetch from database
	if task == nil {
		task, err = h.repo.GetByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, NewErrorResponse("Task not found"))
			return
		}

		// Cache the result for 5 minutes
		if err := h.cache.SetTask(task, 5*time.Minute); err != nil {
			// Log error but don't fail the request
		}
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

// CreateTask handles POST /tasks with cache invalidation
func (h *CachedTaskHandler) CreateTask(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err.Error()))
		return
	}

	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Assignee:    req.Assignee,
	}

	if task.Status == "" {
		task.Status = models.StatusPending
	}

	if err := h.repo.Create(task); err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to create task"))
		return
	}

	// Invalidate tasks list cache
	if err := h.cache.InvalidateTasks(); err != nil {
		// Log error but don't fail the request
	}

	// Update tasks count metric
	if _, total, err := h.repo.GetAll(1, 1, "", ""); err == nil {
		middleware.UpdateTasksCount(float64(total))
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

// UpdateTask handles PUT /tasks/{id} with cache invalidation
func (h *CachedTaskHandler) UpdateTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid task ID"))
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse(err.Error()))
		return
	}

	// Get existing task
	existingTask, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, NewErrorResponse("Task not found"))
		return
	}

	// Update fields if provided
	if req.Title != nil {
		existingTask.Title = *req.Title
	}
	if req.Description != nil {
		existingTask.Description = *req.Description
	}
	if req.Status != nil {
		existingTask.Status = *req.Status
	}
	if req.Assignee != nil {
		existingTask.Assignee = *req.Assignee
	}

	if err := h.repo.Update(existingTask); err != nil {
		c.JSON(http.StatusInternalServerError, NewErrorResponse("Failed to update task"))
		return
	}

	// Invalidate caches
	h.cache.InvalidateTasks()
	h.cache.InvalidateTask(id.String())

	response := TaskResponse{
		ID:          existingTask.ID,
		Title:       existingTask.Title,
		Description: existingTask.Description,
		Status:      existingTask.Status,
		Assignee:    existingTask.Assignee,
		CreatedAt:   existingTask.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   existingTask.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, response)
}

// DeleteTask handles DELETE /tasks/{id} with cache invalidation
func (h *CachedTaskHandler) DeleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, NewErrorResponse("Invalid task ID"))
		return
	}

	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, NewErrorResponse("Task not found"))
		return
	}

	// Invalidate caches
	h.cache.InvalidateTasks()
	h.cache.InvalidateTask(id.String())

	// Update tasks count metric
	if _, total, err := h.repo.GetAll(1, 1, "", ""); err == nil {
		middleware.UpdateTasksCount(float64(total))
	}

	c.JSON(http.StatusNoContent, nil)
}
