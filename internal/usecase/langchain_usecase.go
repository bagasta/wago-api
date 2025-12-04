package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"whatsapp-api/internal/domain/entity"
	"whatsapp-api/internal/domain/repository"
	"whatsapp-api/internal/infrastructure/langchain"
)

type LangchainUseCase struct {
	sessionRepo         repository.SessionRepository
	langchainRepo       repository.LangchainRepository
	langchainClient     *langchain.Client
	defaultLangchainURL string
	defaultParams       map[string]interface{}
}

func NewLangchainUseCase(
	sessionRepo repository.SessionRepository,
	langchainRepo repository.LangchainRepository,
	client *langchain.Client,
	defaultLangchainURL string,
	defaultParams map[string]interface{},
) *LangchainUseCase {
	return &LangchainUseCase{
		sessionRepo:         sessionRepo,
		langchainRepo:       langchainRepo,
		langchainClient:     client,
		defaultLangchainURL: defaultLangchainURL,
		defaultParams:       defaultParams,
	}
}

func (uc *LangchainUseCase) Execute(ctx context.Context, agentID, userMessage, sender string, overrideParams map[string]interface{}) (*entity.LangchainExecution, error) {
	session, err := uc.sessionRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, fmt.Errorf("session not found for agent %s", agentID)
	}

	baseURL := uc.defaultLangchainURL
	if session.LangchainURL.Valid && session.LangchainURL.String != "" {
		baseURL = session.LangchainURL.String
	}
	if baseURL == "" {
		return nil, fmt.Errorf("langchain URL not configured for agent %s", agentID)
	}

	apiKey := session.LangchainAPIKey.String
	if apiKey == "" {
		return nil, fmt.Errorf("langchain API key not set for agent %s", agentID)
	}

	params := uc.defaultParams
	if params == nil {
		params = map[string]interface{}{}
	}
	if len(overrideParams) > 0 {
		merged := make(map[string]interface{}, len(params)+len(overrideParams))
		for k, v := range params {
			merged[k] = v
		}
		for k, v := range overrideParams {
			merged[k] = v
		}
		params = merged
	}

	result, err := uc.langchainClient.Execute(ctx, baseURL, agentID, apiKey, userMessage, sender, params)
	status := sql.NullString{String: "success", Valid: true}
	execTime := sql.NullInt64{Int64: resultDurationMs(result), Valid: true}
	var respBody []byte
	var errMsg sql.NullString

	if result != nil {
		respBody = result.Body
	}

	if err != nil {
		status = sql.NullString{String: "failed", Valid: true}
		errMsg = sql.NullString{String: err.Error(), Valid: true}
	} else if result != nil && result.StatusCode >= 300 {
		status = sql.NullString{String: "failed", Valid: true}
		errMsg = sql.NullString{String: fmt.Sprintf("langchain returned status %d: %s", result.StatusCode, string(result.Body)), Valid: true}
	}

	execution := &entity.LangchainExecution{
		SessionID:         session.ID,
		AgentID:           agentID,
		UserMessage:       sql.NullString{String: userMessage, Valid: userMessage != ""},
		LangchainResponse: respBody,
		ExecutionTimeMs:   execTime,
		Status:            status,
		ErrorMessage:      errMsg,
		CreatedAt:         time.Now(),
	}

	if errCreate := uc.langchainRepo.Create(ctx, execution); errCreate != nil {
		return nil, errCreate
	}

	if status.String == "failed" {
		return execution, fmt.Errorf("%s", errMsg.String)
	}
	return execution, nil
}

func resultDurationMs(result *langchain.ExecuteResult) int64 {
	if result == nil {
		return 0
	}
	return result.Duration.Milliseconds()
}
