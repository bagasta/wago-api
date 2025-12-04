package http

import (
	"whatsapp-api/internal/delivery/http/handler"

	"github.com/gofiber/fiber/v2"
	fiberSwagger "github.com/gofiber/swagger"
)

func NewRouter(app *fiber.App, sessionHandler *handler.SessionHandler, langchainHandler *handler.LangchainHandler) {
	api := app.Group("/api/v1")

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
