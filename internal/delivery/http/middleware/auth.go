package middleware

import (
	"strings"
	"whatsapp-api/internal/domain/repository"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(userRepo repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Missing Authorization header",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid Authorization header format",
			})
		}

		apiKey := parts[1]
		// Debug log
		println("Received API Key:", apiKey)

		user, err := userRepo.GetByAPIKey(c.Context(), apiKey)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Internal Server Error",
			})
		}
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid API Key",
			})
		}

		c.Locals("user", user)
		return c.Next()
	}
}
