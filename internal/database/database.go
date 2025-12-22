package database

import (
	"fmt"
	"log"

	"taheri24.ir/graph1/internal/models"
	"taheri24.ir/graph1/pkg/config"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TaskRepository defines the interface for task database operations
type TaskRepository interface {
	Create(task *models.Task) error
	GetByID(id uuid.UUID) (*models.Task, error)
	GetAll(page, limit int, status, assignee string) ([]models.Task, int64, error)
	Update(task *models.Task) error
	Delete(id uuid.UUID) error
}

type Database struct {
	DB *gorm.DB
}

// Ensure Database implements TaskRepository
var _ TaskRepository = (*Database)(nil)

// Create creates a new task
func (d *Database) Create(task *models.Task) error {
	return d.DB.Create(task).Error
}

// GetByID retrieves a task by ID
func (d *Database) GetByID(id uuid.UUID) (task *models.Task, err error) {
	err = d.DB.First(&task, "id = ?", id).Error
	return task, err
}

// GetAll retrieves tasks with pagination and filtering
func (d *Database) GetAll(page, limit int, status, assignee string) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	offset := (page - 1) * limit
	query := d.DB.Model(&models.Task{})

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if assignee != "" {
		query = query.Where("assignee = ?", assignee)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(offset).Limit(limit).Find(&tasks).Error
	return tasks, total, err
}

// Update updates an existing task
func (d *Database) Update(task *models.Task) error {
	return d.DB.Save(task).Error
}

// Delete deletes a task by ID
func (d *Database) Delete(id uuid.UUID) error {
	return d.DB.Delete(&models.Task{}, "id = ?", id).Error
}

func NewDatabase(cfg *config.Config) (*Database, error) {
	dsn := cfg.GetDatabaseDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&models.Task{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database connection established and migrations completed")

	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *Database) Health() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
