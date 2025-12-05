package handler

import (
	"strings"
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

	if err := h.sessionUC.DeleteSession(c.Context(), req.AgentID); err != nil {
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
