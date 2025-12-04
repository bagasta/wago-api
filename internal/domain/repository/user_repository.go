package repository

import (
	"context"
	"whatsapp-api/internal/domain/entity"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByAPIKey(ctx context.Context, apiKey string) (*entity.User, error)
	GetByID(ctx context.Context, id string) (*entity.User, error)
}
