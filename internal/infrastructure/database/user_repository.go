package database

import (
	"context"
	"database/sql"
	"errors"
	"whatsapp-api/internal/domain/entity"
	"whatsapp-api/internal/domain/repository"

	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (user_id, api_key, created_at, updated_at) 
              VALUES (:user_id, :api_key, :created_at, :updated_at)`

	_, err := r.db.NamedExecContext(ctx, query, user)
	return err
}

func (r *userRepository) GetByAPIKey(ctx context.Context, apiKey string) (*entity.User, error) {
	var user entity.User
	query := `SELECT * FROM users WHERE api_key = $1`

	err := r.db.GetContext(ctx, &user, query, apiKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	var user entity.User
	query := `SELECT * FROM users WHERE user_id = $1`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
