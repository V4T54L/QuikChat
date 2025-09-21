package postgres

import (
	"context"
	"time"

	"chat-app/backend/internal/domain"
	"chat-app/backend/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresMessageRepository struct {
	db *pgxpool.Pool
}

func NewPostgresMessageRepository(db *pgxpool.Pool) repository.MessageRepository {
	return &postgresMessageRepository{db: db}
}

func (r *postgresMessageRepository) Create(ctx context.Context, message *domain.Message) error {
	query := `
        INSERT INTO messages (id, conversation_id, sender_id, content, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := r.db.Exec(ctx, query, message.ID, message.ConversationID, message.SenderID, message.Content, message.CreatedAt)
	return err
}

func (r *postgresMessageRepository) GetByConversationID(ctx context.Context, conversationID string, before time.Time, limit int) ([]*domain.Message, error) {
	query := `
        SELECT m.id, m.conversation_id, m.sender_id, m.content, m.created_at,
               u.username as sender_username, u.profile_pic_url as sender_profile_pic_url
        FROM messages m
        JOIN users u ON m.sender_id = u.id
        WHERE m.conversation_id = $1 AND m.created_at < $2
        ORDER BY m.created_at DESC
        LIMIT $3
    `
	rows, err := r.db.Query(ctx, query, conversationID, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		var msg domain.Message
		msg.Sender = &domain.User{}
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.Content,
			&msg.CreatedAt,
			&msg.Sender.Username,
			&msg.Sender.ProfilePicURL,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}

	// Reverse slice to return messages in ascending order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

