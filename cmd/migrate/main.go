package main

import (
	"context"
	"log"

	"whatsapp-api/internal/infrastructure/database"
	"whatsapp-api/pkg/config"
)

// Migration runner that can be triggered manually:
// go run cmd/migrate/main.go
func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(ctx, db, "migrations"); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations applied successfully")
}
