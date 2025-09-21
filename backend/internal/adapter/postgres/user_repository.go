package postgres

import (
	"context"
	"errors"

	"chat-app/internal/domain"
	"chat-app/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresUserRepository struct {
	db *pgxpool.Pool
}

func NewPostgresUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (id, username, password_hash) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, user.ID, user.Username, user.PasswordHash)
	return err
}

func (r *postgresUserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `SELECT id, username, password_hash, profile_pic_url, created_at, updated_at FROM users WHERE username = $1`
	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.ProfilePicURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `SELECT id, username, password_hash, profile_pic_url, created_at, updated_at FROM users WHERE id = $1`
	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.ProfilePicURL, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `UPDATE users SET username = $1, password_hash = $2, profile_pic_url = $3, updated_at = NOW() WHERE id = $4`
	_, err := r.db.Exec(ctx, query, user.Username, user.PasswordHash, user.ProfilePicURL, user.ID)
	return err
}

