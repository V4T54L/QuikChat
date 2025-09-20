package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"chat-app/models"
	"chat-app/repository"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type postgresFriendshipRepository struct {
	db *sql.DB
}

func NewPostgresFriendshipRepository(db *sql.DB) repository.FriendshipRepository {
	return &postgresFriendshipRepository{db: db}
}

// normalizeUserIDs ensures that userID1 is always less than userID2
func normalizeUserIDs(userID1, userID2 uuid.UUID) (uuid.UUID, uuid.UUID) {
	if userID1.String() > userID2.String() {
		return userID2, userID1
	}
	return userID1, userID2
}

func (r *postgresFriendshipRepository) Create(ctx context.Context, friendship *models.Friendship) error {
	u1, u2 := normalizeUserIDs(friendship.UserID1, friendship.UserID2)
	query := `INSERT INTO friendships (user_id1, user_id2, status) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, u1, u2, friendship.Status)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" { // unique_violation
			return models.ErrFriendRequestExists
		}
		return fmt.Errorf("failed to create friendship: %w", err)
	}
	return nil
}

func (r *postgresFriendshipRepository) UpdateStatus(ctx context.Context, userID1, userID2 uuid.UUID, status models.FriendshipStatus) error {
	u1, u2 := normalizeUserIDs(userID1, userID2)
	query := `UPDATE friendships SET status = $3 WHERE user_id1 = $1 AND user_id2 = $2`
	res, err := r.db.ExecContext(ctx, query, u1, u2, status)
	if err != nil {
		return fmt.Errorf("failed to update friendship status: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return models.ErrFriendRequestNotFound
	}
	return nil
}

func (r *postgresFriendshipRepository) Delete(ctx context.Context, userID1, userID2 uuid.UUID) error {
	u1, u2 := normalizeUserIDs(userID1, userID2)
	query := `DELETE FROM friendships WHERE user_id1 = $1 AND user_id2 = $2`
	res, err := r.db.ExecContext(ctx, query, u1, u2)
	if err != nil {
		return fmt.Errorf("failed to delete friendship: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return models.ErrNotFriends // Or request not found, context dependent
	}
	return nil
}

func (r *postgresFriendshipRepository) Find(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Friendship, error) {
	u1, u2 := normalizeUserIDs(userID1, userID2)
	query := `SELECT user_id1, user_id2, status, created_at FROM friendships WHERE user_id1 = $1 AND user_id2 = $2`
	friendship := &models.Friendship{}
	err := r.db.QueryRowContext(ctx, query, u1, u2).Scan(&friendship.UserID1, &friendship.UserID2, &friendship.Status, &friendship.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrFriendRequestNotFound
		}
		return nil, fmt.Errorf("failed to find friendship: %w", err)
	}
	return friendship, nil
}

func (r *postgresFriendshipRepository) ListByUserID(ctx context.Context, userID uuid.UUID, status models.FriendshipStatus) ([]*models.User, error) {
	query := `
		SELECT u.id, u.username, u.profile_pic_url, u.created_at
		FROM users u
		JOIN friendships f ON (u.id = f.user_id1 OR u.id = f.user_id2)
		WHERE (f.user_id1 = $1 OR f.user_id2 = $1)
		AND u.id != $1
		AND f.status = $2
	`
	rows, err := r.db.QueryContext(ctx, query, userID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to list friends: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.ProfilePicURL, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}
	return users, nil
}

