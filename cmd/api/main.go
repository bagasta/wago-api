package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"whatsapp-api/internal/delivery/http"
	"whatsapp-api/internal/delivery/http/handler"
	"whatsapp-api/internal/domain/entity"
	"whatsapp-api/internal/infrastructure/database"
	"whatsapp-api/internal/infrastructure/langchain"
	"whatsapp-api/internal/infrastructure/whatsapp"
	"whatsapp-api/internal/usecase"
	"whatsapp-api/pkg/config"

	_ "whatsapp-api/docs" // Import generated docs

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// @title WhatsApp API
// @version 1.0
// @description WhatsApp REST API with Go and Fiber
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1
func main() {
	// 1. Load Config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Connect Database
	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 3. Initialize Repositories
	userRepo := database.NewUserRepository(db)
	sessionRepo := database.NewSessionRepository(db)
	messageRepo := database.NewMessageRepository(db)
	langchainRepo := database.NewLangchainRepository(db)

	// Seed default user if not exists
	defaultUserID := "admin"
	seedDefaultUser(userRepo)

	// 4. Initialize Infrastructure
	waManager, err := whatsapp.NewClientManager(db)
	if err != nil {
		log.Fatalf("Failed to initialize WhatsApp manager: %v", err)
	}
	lcTimeout, _ := time.ParseDuration(cfg.Langchain.DefaultTimeout)
	if lcTimeout == 0 {
		lcTimeout = 30 * time.Second
	}
	langchainClient := langchain.NewClient(lcTimeout)

	// 5. Initialize UseCases
	defaultParams := map[string]interface{}{
		"max_steps": 5,
	}
	langchainUC := usecase.NewLangchainUseCase(sessionRepo, langchainRepo, langchainClient, cfg.Langchain.BaseURL, defaultParams)
	sessionUC := usecase.NewSessionUseCase(sessionRepo, messageRepo, waManager, defaultUserID, cfg.Langchain.BaseURL, langchainUC)

	// Initialize existing sessions
	if err := sessionUC.InitializeSessions(context.Background()); err != nil {
		log.Printf("Failed to initialize sessions: %v", err)
	}

	// 6. Initialize Handlers
	sessionHandler := handler.NewSessionHandler(sessionUC)
	langchainHandler := handler.NewLangchainHandler(langchainUC)

	// 7. Initialize Fiber App
	app := fiber.New(fiber.Config{
		AppName: cfg.Server.Name,
	})

	app.Use(logger.New())
	app.Use(recover.New())

	// 8. Setup Router
	http.NewRouter(app, sessionHandler, langchainHandler)

	// 9. Start Server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func seedDefaultUser(userRepo interface{}) {
	// Simple seed for testing
	repo := userRepo.(interface {
		GetByAPIKey(ctx context.Context, apiKey string) (*entity.User, error)
		Create(ctx context.Context, user *entity.User) error
	})

	ctx := context.Background()
	apiKey := "secret"
	user, _ := repo.GetByAPIKey(ctx, apiKey)
	if user == nil {
		log.Println("Seeding default user...")
		newUser := &entity.User{
			UserID:    "admin",
			APIKey:    apiKey,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := repo.Create(ctx, newUser); err != nil {
			log.Printf("Failed to seed user: %v", err)
		} else {
			log.Printf("Default user created. API Key: %s", apiKey)
		}
	}
}
