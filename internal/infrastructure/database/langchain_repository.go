package database

import (
	"context"
	"whatsapp-api/internal/domain/entity"
	"whatsapp-api/internal/domain/repository"

	"github.com/jmoiron/sqlx"
)

type langchainRepository struct {
	db *sqlx.DB
}

func NewLangchainRepository(db *sqlx.DB) repository.LangchainRepository {
	return &langchainRepository{db: db}
}

func (r *langchainRepository) Create(ctx context.Context, execution *entity.LangchainExecution) error {
	query := `INSERT INTO langchain_executions (session_id, agent_id, user_message, langchain_response, execution_time_ms, status, error_message, created_at) 
              VALUES (:session_id, :agent_id, :user_message, :langchain_response, :execution_time_ms, :status, :error_message, :created_at)
			  RETURNING id`

	rows, err := r.db.NamedQueryContext(ctx, query, execution)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&execution.ID)
	}
	return nil
}

func (r *langchainRepository) GetBySessionID(ctx context.Context, sessionID int, limit, offset int) ([]*entity.LangchainExecution, error) {
	var executions []*entity.LangchainExecution
	query := `SELECT * FROM langchain_executions WHERE session_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	err := r.db.SelectContext(ctx, &executions, query, sessionID, limit, offset)
	if err != nil {
		return nil, err
	}

	return executions, nil
}
