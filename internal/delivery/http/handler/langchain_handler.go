package handler

import (
	"encoding/json"

	"whatsapp-api/internal/domain/entity"
	"whatsapp-api/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type LangchainHandler struct {
	uc *usecase.LangchainUseCase
}

func NewLangchainHandler(uc *usecase.LangchainUseCase) *LangchainHandler {
	return &LangchainHandler{uc: uc}
}

type ExecuteLangchainRequest struct {
	AgentID string                 `json:"agentId"`
	Message string                 `json:"message"`
	Sender  string                 `json:"sender,omitempty"`
	Params  map[string]interface{} `json:"params,omitempty"`
}

// Execute godoc
// @Summary Execute Langchain for an agent
// @Description Proxy a user message to the configured Langchain agent and store execution result
// @Tags langchain
// @Accept json
// @Produce json
// @Param request body ExecuteLangchainRequest true "Execution request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /langchain/execute [post]
func (h *LangchainHandler) Execute(c *fiber.Ctx) error {
	var req ExecuteLangchainRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}
	if req.AgentID == "" || req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "agentId and message are required",
		})
	}

	exec, err := h.uc.Execute(c.Context(), req.AgentID, req.Message, req.Sender, req.Params)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"data":    h.presentExecution(exec),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    h.presentExecution(exec),
	})
}

func (h *LangchainHandler) presentExecution(exec *entity.LangchainExecution) fiber.Map {
	if exec == nil {
		return nil
	}
	var parsed interface{}
	if len(exec.LangchainResponse) > 0 {
		_ = json.Unmarshal(exec.LangchainResponse, &parsed)
	}
	return fiber.Map{
		"id":                exec.ID,
		"agentId":           exec.AgentID,
		"sessionId":         exec.SessionID,
		"status":            exec.Status.String,
		"error":             exec.ErrorMessage.String,
		"userMessage":       exec.UserMessage.String,
		"langchainResponse": parsed,
		"rawResponse":       string(exec.LangchainResponse),
		"executionTimeMs":   exec.ExecutionTimeMs.Int64,
		"createdAt":         exec.CreatedAt,
	}
}
