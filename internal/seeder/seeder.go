package seeder

import (
	"log"

	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/models"
)

func Seed(db *database.Database) error {
	tasks := []models.Task{
		{
			Title:       "Complete project setup",
			Description: "Set up the initial project structure and dependencies",
			Status:      "completed",
			Assignee:    "developer",
		},
		{
			Title:       "Implement user authentication",
			Description: "Add user login and registration functionality",
			Status:      "in_progress",
			Assignee:    "developer",
		},
		{
			Title:       "Write unit tests",
			Description: "Create comprehensive unit tests for all modules",
			Status:      "pending",
			Assignee:    "tester",
		},
		{
			Title:       "Deploy to production",
			Description: "Deploy the application to the production environment",
			Status:      "pending",
			Assignee:    "devops",
		},
	}

	for _, task := range tasks {
		err := db.DB.Exec(`
			INSERT INTO tasks (title, description, status, assignee)
			VALUES (?, ?, ?, ?)
			ON CONFLICT (id) DO NOTHING
		`, task.Title, task.Description, task.Status, task.Assignee).Error
		if err != nil {
			log.Printf("Error seeding task %s: %v", task.Title, err)
			return err
		}
	}

	log.Println("Database seeded successfully")
	return nil
}
