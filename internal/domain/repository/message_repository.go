package repository

import (
	"context"
	"whatsapp-api/internal/domain/entity"
)

type MessageRepository interface {
	Create(ctx context.Context, message *entity.Message) error
	GetBySessionID(ctx context.Context, sessionID int, limit, offset int) ([]*entity.Message, error)
	CountByAgentAndDirection(ctx context.Context, agentID string, direction string) (int, error)
}
