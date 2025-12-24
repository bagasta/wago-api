package handler

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"whatsapp-api/internal/domain/entity"
	"whatsapp-api/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type SessionHandler struct {
	sessionUC *usecase.SessionUseCase
}

func NewSessionHandler(sessionUC *usecase.SessionUseCase) *SessionHandler {
	return &SessionHandler{sessionUC: sessionUC}
}

type CreateSessionRequest struct {
	AgentID      string `json:"agentId"`
	AgentName    string `json:"agentName"`
	APIKey       string `json:"apiKey"`
	LangchainURL string `json:"langchainUrl"`
}

// CreateSession godoc
// @Summary Create a new WhatsApp session
// @Description Create a new WhatsApp session and generate QR code
// @Tags sessions
// @Accept json
// @Produce json
// @Param request body CreateSessionRequest true "Session Creation Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /sessions/create [post]
func (h *SessionHandler) CreateSession(c *fiber.Ctx) error {
	var req CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.AgentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "AgentID is required",
		})
	}

	session, err := h.sessionUC.CreateSession(c.Context(), req.AgentID, req.AgentName, req.APIKey, req.LangchainURL)
	if err != nil {
		if strings.Contains(err.Error(), "session already exists") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"error":   err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Session created successfully",
		"data": fiber.Map{
			"sessionId":       session.ID,
			"agentId":         session.AgentID,
			"qrCode":          dataURLFromBase64(session.QRCodeBase64.String),
			"qrCodeBase64":    stripDataURLPrefix(session.QRCodeBase64.String),
			"status":          session.Status,
			"lastGeneratedAt": session.LastQRGeneratedAt.Time,
		},
	})
}

// CreateTestSession godoc
// @Summary Create a mock connected session for testing (TEST ONLY)
// @Description Creates a session that bypasses QR scanning and appears as 'connected'. Only available in development/testing environments.
// @Tags sessions
// @Accept json
// @Produce json
// @Param request body CreateSessionRequest true "Session Creation Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /sessions/create-test [post]
func (h *SessionHandler) CreateTestSession(c *fiber.Ctx) error {
	// Only enable in test/dev environment
	env := os.Getenv("APP_ENV")
	if env != "development" && env != "testing" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"error":   "Endpoint only available in test/development mode",
		})
	}

	var req CreateSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.AgentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "AgentID is required",
		})
	}

	// Create a mock session directly in the database
	ctx := c.Context()
	session, err := h.createMockConnectedSession(ctx, req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"success": false,
				"error":   "Session already exists",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Test session created successfully (mock connected)",
		"data": fiber.Map{
			"sessionId":   session.ID,
			"agentId":     session.AgentID,
			"status":      session.Status,
			"phoneNumber": session.PhoneNumber.String,
			"connectedAt": session.ConnectedAt.Time,
		},
	})
}

func (h *SessionHandler) createMockConnectedSession(ctx context.Context, req CreateSessionRequest) (*entity.Session, error) {
	// Check if session already exists
	existing, _ := h.sessionUC.GetSession(ctx, req.AgentID)
	if existing != nil {
		return nil, errors.New("session already exists with this agentId")
	}

	// This endpoint requires direct database access which we don't have in the handler.
	// For automated testing, please use the SQL seed script (tests/seed_test_sessions.sql)
	// or set up sessions via database directly before load testing.
	return nil, errors.New("please use the SQL seed script (tests/seed_test_sessions.sql) to create test sessions")
}

type AgentRequest struct {
	AgentID string `json:"agentId"`
}

// GetSessionStatus godoc
// @Summary Get session status
// @Description Get the status of a WhatsApp session
// @Tags sessions
// @Accept json
// @Produce json
// @Param agentId query string true "Agent ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /sessions/status [get]
func (h *SessionHandler) GetSessionStatus(c *fiber.Ctx) error {
	var req AgentRequest
	if err := c.BodyParser(&req); err != nil || req.AgentID == "" {
		// Try query param
		req.AgentID = c.Query("agentId")
	}

	if req.AgentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "AgentID is required",
		})
	}

	session, err := h.sessionUC.GetSession(c.Context(), req.AgentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Session not found",
		})
	}
	if session == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Session not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"agentId":           session.AgentID,
			"status":            session.Status,
			"phoneNumber":       session.PhoneNumber.String,
			"connectedAt":       session.ConnectedAt.Time,
			"qrCode":            dataURLFromBase64(session.QRCodeBase64.String),
			"qrCodeBase64":      stripDataURLPrefix(session.QRCodeBase64.String),
			"lastQrGeneratedAt": session.LastQRGeneratedAt.Time,
		},
	})
}

func stripDataURLPrefix(raw string) string {
	const prefix = "data:image/png;base64,"
	if len(raw) >= len(prefix) && raw[:len(prefix)] == prefix {
		return raw[len(prefix):]
	}
	return raw
}

func dataURLFromBase64(raw string) string {
	if raw == "" {
		return ""
	}
	return "data:image/png;base64," + stripDataURLPrefix(raw)
}

// DeleteSession godoc
// @Summary Delete a session
// @Description Delete a WhatsApp session
// @Tags sessions
// @Accept json
// @Produce json
// @Param request body AgentRequest true "Agent Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /sessions/delete [delete]
func (h *SessionHandler) DeleteSession(c *fiber.Ctx) error {
	var req AgentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}
	if req.AgentID == "" {
		req.AgentID = c.Query("agentId")
	}
	if req.AgentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "AgentID is required",
		})
	}

	if err := h.sessionUC.DeleteSession(c.Context(), req.AgentID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"error":   "Session not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Session deleted successfully",
	})
}

// GetSessionDetail godoc
// @Summary Get session details
// @Description Get detailed information about a WhatsApp session
// @Tags sessions
// @Accept json
// @Produce json
// @Param agentId query string true "Agent ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /sessions/detail [get]
func (h *SessionHandler) GetSessionDetail(c *fiber.Ctx) error {
	var req AgentRequest
	if err := c.BodyParser(&req); err != nil || req.AgentID == "" {
		req.AgentID = c.Query("agentId")
	}

	session, err := h.sessionUC.GetSession(c.Context(), req.AgentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Session not found",
		})
	}
	if session == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Session not found",
		})
	}

	stats := h.sessionUC.GetMessageStats(c.Context(), req.AgentID)

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"session": session,
			"stats": fiber.Map{
				"incoming":  stats.Incoming,
				"responded": stats.Responded,
			},
		},
	})
}

// ReconnectSession godoc
// @Summary Reconnect a session
// @Description Reconnect a WhatsApp session
// @Tags sessions
// @Accept json
// @Produce json
// @Param request body AgentRequest true "Agent Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /sessions/reconnect [post]
func (h *SessionHandler) ReconnectSession(c *fiber.Ctx) error {
	var req AgentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	if req.AgentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "AgentID is required",
		})
	}

	session, err := h.sessionUC.ReconnectSession(c.Context(), req.AgentID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Session reconnected successfully",
		"data": fiber.Map{
			"sessionId":       session.ID,
			"agentId":         session.AgentID,
			"qrCode":          dataURLFromBase64(session.QRCodeBase64.String),
			"qrCodeBase64":    stripDataURLPrefix(session.QRCodeBase64.String),
			"status":          session.Status,
			"lastGeneratedAt": session.LastQRGeneratedAt.Time,
		},
	})
}
