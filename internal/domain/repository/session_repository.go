package repository

import (
	"context"
	"whatsapp-api/internal/domain/entity"
)

type SessionRepository interface {
	Create(ctx context.Context, session *entity.Session) error
	Update(ctx context.Context, session *entity.Session) error
	Delete(ctx context.Context, agentID string) error
	GetByAgentID(ctx context.Context, agentID string) (*entity.Session, error)
	GetByUserIDAndAgentID(ctx context.Context, userID, agentID string) (*entity.Session, error)
	GetAllSessions(ctx context.Context) ([]*entity.Session, error)
}
