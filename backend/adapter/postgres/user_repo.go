package postgres

import (
	"chat-app/backend/models"
	"chat-app/backend/repository"
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type postgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) repository.UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (id, username, password_hash, profile_pic_url) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.PasswordHash, user.ProfilePicURL)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" { // unique_violation
			return models.ErrUsernameTaken
		}
		return err
	}
	return nil
}

func (r *postgresUserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `SELECT id, username, password_hash, profile_pic_url, created_at FROM users WHERE LOWER(username) = $1`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, strings.ToLower(username)).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.ProfilePicURL, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `SELECT id, username, password_hash, profile_pic_url, created_at FROM users WHERE id = $1`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.ProfilePicURL, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) Update(ctx context.Context, user *models.User) error {
	query := `UPDATE users SET username = $1, password_hash = $2, profile_pic_url = $3 WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, user.Username, user.PasswordHash, user.ProfilePicURL, user.ID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return models.ErrUsernameTaken
		}
		return err
	}
	return nil
}

