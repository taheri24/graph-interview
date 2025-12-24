package seeder

import (
	"testing"

	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/models"
	"taheri24.ir/graph1/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeed(t *testing.T) {
	cfg := config.NewTestConfig()
	db, err := database.NewDatabase(cfg)
	require.NoError(t, err)

	err = database.Migrate(db.DB)
	require.NoError(t, err)

	err = Seed(db)
	assert.NoError(t, err)

	var count int64
	db.DB.Model(&models.Task{}).Count(&count)
	assert.Equal(t, int64(4), count)

	var tasks []models.Task
	db.DB.Find(&tasks)
	assert.Len(t, tasks, 4)

	// Check first task
	assert.Equal(t, "Complete project setup", tasks[0].Title)
	assert.Equal(t, "Set up the initial project structure and dependencies", tasks[0].Description)
	assert.Equal(t, models.StatusCompleted, tasks[0].Status)
	assert.Equal(t, "developer", tasks[0].Assignee)

	// Check second task
	assert.Equal(t, "Implement user authentication", tasks[1].Title)
	assert.Equal(t, "Add user login and registration functionality", tasks[1].Description)
	assert.Equal(t, models.StatusInProgress, tasks[1].Status)
	assert.Equal(t, "developer", tasks[1].Assignee)

	// Check third task
	assert.Equal(t, "Write unit tests", tasks[2].Title)
	assert.Equal(t, "Create comprehensive unit tests for all modules", tasks[2].Description)
	assert.Equal(t, models.StatusPending, tasks[2].Status)
	assert.Equal(t, "tester", tasks[2].Assignee)

	// Check fourth task
	assert.Equal(t, "Deploy to production", tasks[3].Title)
	assert.Equal(t, "Deploy the application to the production environment", tasks[3].Description)
	assert.Equal(t, models.StatusPending, tasks[3].Status)
	assert.Equal(t, "devops", tasks[3].Assignee)
}
