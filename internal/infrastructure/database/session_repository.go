package database

import (
	"context"
	"database/sql"
	"errors"
	"whatsapp-api/internal/domain/entity"
	"whatsapp-api/internal/domain/repository"

	"github.com/jmoiron/sqlx"
)

type sessionRepository struct {
	db *sqlx.DB
}

func NewSessionRepository(db *sqlx.DB) repository.SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *entity.Session) error {
	query := `INSERT INTO sessions (user_id, agent_id, agent_name, phone_number, qr_code, qr_code_base64, session_data, status, langchain_url, langchain_api_key, last_qr_generated_at, connected_at, disconnected_at, created_at, updated_at) 
              VALUES (:user_id, :agent_id, :agent_name, :phone_number, :qr_code, :qr_code_base64, :session_data, :status, :langchain_url, :langchain_api_key, :last_qr_generated_at, :connected_at, :disconnected_at, :created_at, :updated_at)
			  RETURNING id`

	rows, err := r.db.NamedQueryContext(ctx, query, session)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&session.ID)
	}
	return nil
}

func (r *sessionRepository) Update(ctx context.Context, session *entity.Session) error {
	query := `UPDATE sessions SET 
              agent_name=:agent_name, phone_number=:phone_number, qr_code=:qr_code, qr_code_base64=:qr_code_base64, 
              session_data=:session_data, status=:status, langchain_url=:langchain_url, langchain_api_key=:langchain_api_key,
              last_qr_generated_at=:last_qr_generated_at, connected_at=:connected_at, disconnected_at=:disconnected_at, 
              updated_at=:updated_at
              WHERE id=:id`

	_, err := r.db.NamedExecContext(ctx, query, session)
	return err
}

func (r *sessionRepository) Delete(ctx context.Context, agentID string) error {
	query := `DELETE FROM sessions WHERE agent_id = $1`
	_, err := r.db.ExecContext(ctx, query, agentID)
	return err
}

func (r *sessionRepository) GetByAgentID(ctx context.Context, agentID string) (*entity.Session, error) {
	var session entity.Session
	query := `SELECT * FROM sessions WHERE agent_id = $1`

	err := r.db.GetContext(ctx, &session, query, agentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &session, nil
}

func (r *sessionRepository) GetByUserIDAndAgentID(ctx context.Context, userID, agentID string) (*entity.Session, error) {
	var session entity.Session
	query := `SELECT * FROM sessions WHERE user_id = $1 AND agent_id = $2`

	err := r.db.GetContext(ctx, &session, query, userID, agentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &session, nil
}

func (r *sessionRepository) GetAllSessions(ctx context.Context) ([]*entity.Session, error) {
	var sessions []*entity.Session
	query := `SELECT * FROM sessions`

	err := r.db.SelectContext(ctx, &sessions, query)
	if err != nil {
		return nil, err
	}

	return sessions, nil
}
