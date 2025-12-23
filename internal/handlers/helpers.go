package handlers

import (
	"taheri24.ir/graph1/internal/models"
)

// tasksToResponses converts models.Task to TaskResponse
func tasksToResponses(tasks []models.Task) []TaskResponse {
	responses := make([]TaskResponse, len(tasks))
	for i, task := range tasks {
		responses[i] = TaskResponse{
			ID:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			Status:      task.Status,
			Assignee:    task.Assignee,
			CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return responses
}

// filterTasksByStatus filters tasks by status
func filterTasksByStatus(tasks []models.Task, status models.TaskStatus) []models.Task {
	var filtered []models.Task
	for _, task := range tasks {
		if task.Status == status {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// filterTasksByAssignee filters tasks by assignee
func filterTasksByAssignee(tasks []models.Task, assignee string) []models.Task {
	var filtered []models.Task
	for _, task := range tasks {
		if task.Assignee == assignee {
			filtered = append(filtered, task)
		}
	}
	return filtered
}
