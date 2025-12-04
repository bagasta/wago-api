package database

import (
	"context"
	"whatsapp-api/internal/domain/entity"
	"whatsapp-api/internal/domain/repository"

	"github.com/jmoiron/sqlx"
)

type messageRepository struct {
	db *sqlx.DB
}

func NewMessageRepository(db *sqlx.DB) repository.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, message *entity.Message) error {
	query := `INSERT INTO messages (session_id, agent_id, message_id, from_number, to_number, message_text, message_type, direction, status, metadata, created_at) 
              VALUES (:session_id, :agent_id, :message_id, :from_number, :to_number, :message_text, :message_type, :direction, :status, :metadata, :created_at)
			  RETURNING id`

	rows, err := r.db.NamedQueryContext(ctx, query, message)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.Scan(&message.ID)
	}
	return nil
}

func (r *messageRepository) GetBySessionID(ctx context.Context, sessionID int, limit, offset int) ([]*entity.Message, error) {
	var messages []*entity.Message
	query := `SELECT * FROM messages WHERE session_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`

	err := r.db.SelectContext(ctx, &messages, query, sessionID, limit, offset)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *messageRepository) CountByAgentAndDirection(ctx context.Context, agentID string, direction string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM messages WHERE agent_id = $1 AND direction = $2`
	err := r.db.GetContext(ctx, &count, query, agentID, direction)
	return count, err
}
