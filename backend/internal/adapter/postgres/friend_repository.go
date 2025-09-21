package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"chat-app/internal/domain"
	"chat-app/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresFriendRepository struct {
	db *pgxpool.Pool
}

func NewPostgresFriendRepository(db *pgxpool.Pool) repository.FriendRepository {
	return &postgresFriendRepository{db: db}
}

func (r *postgresFriendRepository) CreateRequest(ctx context.Context, req *domain.FriendRequest) error {
	query := `
        INSERT INTO friend_requests (id, sender_id, receiver_id, status, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
	_, err := r.db.Exec(ctx, query, req.ID, req.SenderID, req.ReceiverID, req.Status, req.CreatedAt, req.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return repository.ErrFriendRequestExists
		}
		return fmt.Errorf("failed to create friend request: %w", err)
	}
	return nil
}

func (r *postgresFriendRepository) GetRequestByID(ctx context.Context, id string) (*domain.FriendRequest, error) {
	query := `
        SELECT id, sender_id, receiver_id, status, created_at, updated_at
        FROM friend_requests
        WHERE id = $1
    `
	req := &domain.FriendRequest{}
	err := r.db.QueryRow(ctx, query, id).Scan(&req.ID, &req.SenderID, &req.ReceiverID, &req.Status, &req.CreatedAt, &req.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get friend request by id: %w", err)
	}
	return req, nil
}

func (r *postgresFriendRepository) UpdateRequestStatus(ctx context.Context, id, status string) error {
	query := `
        UPDATE friend_requests
        SET status = $1, updated_at = $2
        WHERE id = $3
    `
	cmdTag, err := r.db.Exec(ctx, query, status, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("failed to update friend request status: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *postgresFriendRepository) AreFriends(ctx context.Context, userID1, userID2 string) (bool, error) {
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}
	query := `
        SELECT EXISTS (
            SELECT 1 FROM friendships WHERE user_id1 = $1 AND user_id2 = $2
        )
    `
	var exists bool
	err := r.db.QueryRow(ctx, query, userID1, userID2).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check friendship: %w", err)
	}
	return exists, nil
}

func (r *postgresFriendRepository) AddFriendship(ctx context.Context, userID1, userID2 string) error {
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}
	query := `
        INSERT INTO friendships (user_id1, user_id2)
        VALUES ($1, $2)
        ON CONFLICT (user_id1, user_id2) DO NOTHING
    `
	_, err := r.db.Exec(ctx, query, userID1, userID2)
	if err != nil {
		return fmt.Errorf("failed to add friendship: %w", err)
	}
	return nil
}

func (r *postgresFriendRepository) RemoveFriendship(ctx context.Context, userID1, userID2 string) error {
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}
	query := `
        DELETE FROM friendships
        WHERE user_id1 = $1 AND user_id2 = $2
    `
	_, err := r.db.Exec(ctx, query, userID1, userID2)
	if err != nil {
		return fmt.Errorf("failed to remove friendship: %w", err)
	}
	return nil
}

func (r *postgresFriendRepository) GetPendingRequests(ctx context.Context, userID string) ([]*domain.FriendRequest, error) {
	query := `
        SELECT fr.id, fr.sender_id, fr.receiver_id, fr.status, fr.created_at, fr.updated_at,
               s.username as sender_username, s.profile_pic_url as sender_profile_pic_url,
               rc.username as receiver_username, rc.profile_pic_url as receiver_profile_pic_url
        FROM friend_requests fr
        JOIN users s ON fr.sender_id = s.id
        JOIN users rc ON fr.receiver_id = rc.id
        WHERE (fr.receiver_id = $1 OR fr.sender_id = $1) AND fr.status = 'pending'
        ORDER BY fr.created_at DESC
    `
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending requests: %w", err)
	}
	defer rows.Close()

	var requests []*domain.FriendRequest
	for rows.Next() {
		req := &domain.FriendRequest{
			Sender:   &domain.User{},
			Receiver: &domain.User{},
		}
		err := rows.Scan(
			&req.ID, &req.SenderID, &req.ReceiverID, &req.Status, &req.CreatedAt, &req.UpdatedAt,
			&req.Sender.Username, &req.Sender.ProfilePicURL,
			&req.Receiver.Username, &req.Receiver.ProfilePicURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending request: %w", err)
		}
		req.Sender.ID = req.SenderID
		req.Receiver.ID = req.ReceiverID
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *postgresFriendRepository) GetFriendsByUserID(ctx context.Context, userID string) ([]*domain.User, error) {
	query := `
        SELECT u.id, u.username, u.profile_pic_url, u.created_at
        FROM friendships f
        JOIN users u ON u.id = CASE WHEN f.user_id1 = $1 THEN f.user_id2 ELSE f.user_id1 END
        WHERE f.user_id1 = $1 OR f.user_id2 = $1
        ORDER BY u.username
    `
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get friends by user id: %w", err)
	}
	defer rows.Close()

	var friends []*domain.User
	for rows.Next() {
		friend := &domain.User{}
		var createdAt sql.NullTime
		err := rows.Scan(&friend.ID, &friend.Username, &friend.ProfilePicURL, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan friend: %w", err)
		}
		if createdAt.Valid {
			friend.CreatedAt = createdAt.Time
		}
		friends = append(friends, friend)
	}
	return friends, nil
}

func (r *postgresFriendRepository) HasPendingRequest(ctx context.Context, userID1, userID2 string) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1 FROM friend_requests
            WHERE ((sender_id = $1 AND receiver_id = $2) OR (sender_id = $2 AND receiver_id = $1))
            AND status = 'pending'
        )
    `
	var exists bool
	err := r.db.QueryRow(ctx, query, userID1, userID2).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check for pending request: %w", err)
	}
	return exists, nil
}

