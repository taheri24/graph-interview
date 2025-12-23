package main

import (
	"log"

	"taheri24.ir/graph1/internal/database"
	"taheri24.ir/graph1/internal/seeder"
	"taheri24.ir/graph1/pkg/config"
	"taheri24.ir/graph1/pkg/utils"
)

func main() {
	cfg := config.Load()

	db := utils.Must(database.NewDatabase(cfg))
	defer db.Close()

	if err := db.Health(); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}

	if err := database.Migrate(db.DB); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	if err := seeder.Seed(db); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	log.Println("Seeding completed successfully")
}
