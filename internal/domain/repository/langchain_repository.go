package repository

import (
	"context"
	"whatsapp-api/internal/domain/entity"
)

type LangchainRepository interface {
	Create(ctx context.Context, execution *entity.LangchainExecution) error
	GetBySessionID(ctx context.Context, sessionID int, limit, offset int) ([]*entity.LangchainExecution, error)
}
