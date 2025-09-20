package postgres

import (
	"chat-app/backend/models"
	"chat-app/backend/repository"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type postgresSessionRepository struct {
	db *sql.DB
}

func NewPostgresSessionRepository(db *sql.DB) repository.SessionRepository {
	return &postgresSessionRepository{db: db}
}

func (r *postgresSessionRepository) Create(ctx context.Context, session *models.Session) error {
	query := `
        INSERT INTO sessions (refresh_token, user_id, expires_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (refresh_token) DO UPDATE SET
        expires_at = EXCLUDED.expires_at`
	_, err := r.db.ExecContext(ctx, query, session.RefreshToken, session.UserID, session.ExpiresAt)
	return err
}

func (r *postgresSessionRepository) Find(ctx context.Context, refreshToken uuid.UUID) (*models.Session, error) {
	query := `SELECT refresh_token, user_id, expires_at FROM sessions WHERE refresh_token = $1`
	session := &models.Session{}
	err := r.db.QueryRowContext(ctx, query, refreshToken).Scan(&session.RefreshToken, &session.UserID, &session.ExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrSessionNotFound
		}
		return nil, err
	}
	return session, nil
}

func (r *postgresSessionRepository) Delete(ctx context.Context, refreshToken uuid.UUID) error {
	query := `DELETE FROM sessions WHERE refresh_token = $1`
	_, err := r.db.ExecContext(ctx, query, refreshToken)
	return err
}

func (r *postgresSessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
