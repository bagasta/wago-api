package http

import (
	"whatsapp-api/internal/delivery/http/handler"

	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/gofiber/swagger"
)

func NewRouter(app *fiber.App, sessionHandler *handler.SessionHandler, langchainHandler *handler.LangchainHandler) {
	// Simple health checks for uptime monitoring and load balancers
	app.Get("/health", healthHandler)

	api := app.Group("/api/v1")
	api.Get("/health", healthHandler)

	sessions := api.Group("/sessions")
	sessions.Post("/create", sessionHandler.CreateSession)
	sessions.Get("/status", sessionHandler.GetSessionStatus)
	sessions.Delete("/delete", sessionHandler.DeleteSession)
	sessions.Get("/detail", sessionHandler.GetSessionDetail)
	sessions.Post("/reconnect", sessionHandler.ReconnectSession)
	// Add other routes here

	langchain := api.Group("/langchain")
	langchain.Post("/execute", langchainHandler.Execute)

	// Swagger
	app.Get("/swagger/*", fiberSwagger.HandlerDefault)
}

func healthHandler(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
}
