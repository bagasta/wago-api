package entity

import (
	"database/sql"
	"time"
)

type Session struct {
	ID                int            `json:"id" db:"id"`
	UserID            string         `json:"userId" db:"user_id"`
	AgentID           string         `json:"agentId" db:"agent_id"`
	AgentName         sql.NullString `json:"agentName" db:"agent_name"`
	PhoneNumber       sql.NullString `json:"phoneNumber" db:"phone_number"`
	QRCode            sql.NullString `json:"qrCode" db:"qr_code"`
	QRCodeBase64      sql.NullString `json:"qrCodeBase64" db:"qr_code_base64"`
	SessionData       []byte         `json:"sessionData" db:"session_data"` // JSONB stored as byte array
	Status            string         `json:"status" db:"status"`
	LangchainURL      sql.NullString `json:"langchainUrl" db:"langchain_url"`
	LangchainAPIKey   sql.NullString `json:"langchainApiKey" db:"langchain_api_key"`
	LastQRGeneratedAt sql.NullTime   `json:"lastQrGeneratedAt" db:"last_qr_generated_at"`
	ConnectedAt       sql.NullTime   `json:"connectedAt" db:"connected_at"`
	DisconnectedAt    sql.NullTime   `json:"disconnectedAt" db:"disconnected_at"`
	CreatedAt         time.Time      `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time      `json:"updatedAt" db:"updated_at"`
}
