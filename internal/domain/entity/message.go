package entity

import (
	"database/sql"
	"time"
)

type Message struct {
	ID          int            `json:"id" db:"id"`
	SessionID   int            `json:"sessionId" db:"session_id"`
	AgentID     string         `json:"agentId" db:"agent_id"`
	MessageID   sql.NullString `json:"messageId" db:"message_id"`
	FromNumber  sql.NullString `json:"fromNumber" db:"from_number"`
	ToNumber    sql.NullString `json:"toNumber" db:"to_number"`
	MessageText sql.NullString `json:"messageText" db:"message_text"`
	MessageType sql.NullString `json:"messageType" db:"message_type"`
	Direction   sql.NullString `json:"direction" db:"direction"`
	Status      sql.NullString `json:"status" db:"status"`
	Metadata    []byte         `json:"metadata" db:"metadata"` // JSONB
	CreatedAt   time.Time      `json:"createdAt" db:"created_at"`
}
