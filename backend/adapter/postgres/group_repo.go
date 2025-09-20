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

type postgresGroupRepository struct {
	db *sql.DB
}

func NewPostgresGroupRepository(db *sql.DB) repository.GroupRepository {
	return &postgresGroupRepository{db: db}
}

func (r *postgresGroupRepository) Create(ctx context.Context, group *models.Group) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `INSERT INTO groups (id, handle, name, photo_url, owner_id) VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.ExecContext(ctx, query, group.ID, group.Handle, group.Name, group.PhotoURL, group.OwnerID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" { // unique_violation
			return models.ErrGroupHandleTaken
		}
		return fmt.Errorf("failed to create group: %w", err)
	}

	memberQuery := `INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)`
	_, err = tx.ExecContext(ctx, memberQuery, group.ID, group.OwnerID)
	if err != nil {
		return fmt.Errorf("failed to add owner as member: %w", err)
	}

	return tx.Commit()
}

func (r *postgresGroupRepository) Update(ctx context.Context, group *models.Group) error {
	query := `UPDATE groups SET name = $2, photo_url = $3, owner_id = $4 WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, group.ID, group.Name, group.PhotoURL, group.OwnerID)
	if err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return models.ErrGroupNotFound
	}
	return nil
}

func (r *postgresGroupRepository) Delete(ctx context.Context, groupID uuid.UUID) error {
	query := `DELETE FROM groups WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, groupID)
	return err
}

func (r *postgresGroupRepository) FindByID(ctx context.Context, groupID uuid.UUID) (*models.Group, error) {
	query := `SELECT id, handle, name, photo_url, owner_id, created_at FROM groups WHERE id = $1`
	group := &models.Group{}
	err := r.db.QueryRowContext(ctx, query, groupID).Scan(&group.ID, &group.Handle, &group.Name, &group.PhotoURL, &group.OwnerID, &group.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrGroupNotFound
		}
		return nil, fmt.Errorf("failed to find group by id: %w", err)
	}
	return group, nil
}

func (r *postgresGroupRepository) FindByHandle(ctx context.Context, handle string) (*models.Group, error) {
	query := `SELECT id, handle, name, photo_url, owner_id, created_at FROM groups WHERE LOWER(handle) = LOWER($1)`
	group := &models.Group{}
	err := r.db.QueryRowContext(ctx, query, handle).Scan(&group.ID, &group.Handle, &group.Name, &group.PhotoURL, &group.OwnerID, &group.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrGroupNotFound
		}
		return nil, fmt.Errorf("failed to find group by handle: %w", err)
	}
	return group, nil
}

func (r *postgresGroupRepository) FuzzySearchByHandle(ctx context.Context, query string, limit int) ([]*models.Group, error) {
	// Note: For true fuzzy search, extensions like pg_trgm are better. This is a simple LIKE search.
	sqlQuery := `
		SELECT id, handle, name, photo_url, owner_id, created_at
		FROM groups
		WHERE LOWER(handle) LIKE LOWER($1)
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, sqlQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search groups: %w", err)
	}
	defer rows.Close()

	var groups []*models.Group
	for rows.Next() {
		group := &models.Group{}
		if err := rows.Scan(&group.ID, &group.Handle, &group.Name, &group.PhotoURL, &group.OwnerID, &group.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan group row: %w", err)
		}
		groups = append(groups, group)
	}
	return groups, nil
}

func (r *postgresGroupRepository) AddMember(ctx context.Context, member *models.GroupMember) error {
	query := `INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, member.GroupID, member.UserID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" { // unique_violation
			return models.ErrAlreadyGroupMember
		}
		return fmt.Errorf("failed to add group member: %w", err)
	}
	return nil
}

func (r *postgresGroupRepository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	query := `DELETE FROM group_members WHERE group_id = $1 AND user_id = $2`
	res, err := r.db.ExecContext(ctx, query, groupID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove group member: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return models.ErrNotGroupMember
	}
	return nil
}

func (r *postgresGroupRepository) FindMember(ctx context.Context, groupID, userID uuid.UUID) (*models.GroupMember, error) {
	query := `SELECT group_id, user_id, joined_at FROM group_members WHERE group_id = $1 AND user_id = $2`
	member := &models.GroupMember{}
	err := r.db.QueryRowContext(ctx, query, groupID, userID).Scan(&member.GroupID, &member.UserID, &member.JoinedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNotGroupMember
		}
		return nil, fmt.Errorf("failed to find group member: %w", err)
	}
	return member, nil
}

func (r *postgresGroupRepository) ListMembers(ctx context.Context, groupID uuid.UUID) ([]*models.User, error) {
	query := `
		SELECT u.id, u.username, u.profile_pic_url, u.created_at
		FROM users u
		JOIN group_members gm ON u.id = gm.user_id
		WHERE gm.group_id = $1
		ORDER BY gm.joined_at
	`
	rows, err := r.db.QueryContext(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to list group members: %w", err)
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

func (r *postgresGroupRepository) GetOldestMember(ctx context.Context, groupID uuid.UUID) (*models.User, error) {
	query := `
		SELECT u.id, u.username, u.profile_pic_url, u.created_at
		FROM users u
		JOIN group_members gm ON u.id = gm.user_id
		WHERE gm.group_id = $1
		ORDER BY gm.joined_at ASC
		LIMIT 1
	`
	user := &models.User{}
	err := r.db.QueryRowContext(ctx, query, groupID).Scan(&user.ID, &user.Username, &user.ProfilePicURL, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrGroupNotFound // Or no members, but group should have at least one
		}
		return nil, fmt.Errorf("failed to get oldest member: %w", err)
	}
	return user, nil
}

