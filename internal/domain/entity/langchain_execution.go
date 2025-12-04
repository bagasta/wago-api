package entity

import (
	"database/sql"
	"time"
)

type LangchainExecution struct {
	ID                int            `json:"id" db:"id"`
	SessionID         int            `json:"sessionId" db:"session_id"`
	AgentID           string         `json:"agentId" db:"agent_id"`
	UserMessage       sql.NullString `json:"userMessage" db:"user_message"`
	LangchainResponse []byte         `json:"langchainResponse" db:"langchain_response"` // JSONB
	ExecutionTimeMs   sql.NullInt64  `json:"executionTimeMs" db:"execution_time_ms"`
	Status            sql.NullString `json:"status" db:"status"`
	ErrorMessage      sql.NullString `json:"errorMessage" db:"error_message"`
	CreatedAt         time.Time      `json:"createdAt" db:"created_at"`
}
