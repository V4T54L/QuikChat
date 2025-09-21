package postgres

import (
	"chat-app/internal/domain"
	"chat-app/internal/repository"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresGroupRepository struct {
	db *pgxpool.Pool
}

func NewPostgresGroupRepository(db *pgxpool.Pool) repository.GroupRepository {
	return &postgresGroupRepository{db: db}
}

func (r *postgresGroupRepository) Create(ctx context.Context, group *domain.Group) error {
	query := `INSERT INTO groups (id, handle, name, owner_id, profile_pic_url)
              VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(ctx, query, group.ID, group.Handle, group.Name, group.OwnerID, group.ProfilePicURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return repository.ErrGroupHandleExists
		}
		return fmt.Errorf("failed to create group: %w", err)
	}
	return nil
}

func (r *postgresGroupRepository) CreateMember(ctx context.Context, member *domain.GroupMember) error {
	query := `INSERT INTO group_members (group_id, user_id, role) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(ctx, query, member.GroupID, member.UserID, member.Role)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return repository.ErrGroupMemberExists
		}
		return fmt.Errorf("failed to create group member: %w", err)
	}
	return nil
}

func (r *postgresGroupRepository) GetByID(ctx context.Context, id string) (*domain.Group, error) {
	query := `SELECT id, handle, name, owner_id, profile_pic_url, created_at, updated_at
              FROM groups WHERE id = $1`
	var group domain.Group
	err := r.db.QueryRow(ctx, query, id).Scan(
		&group.ID, &group.Handle, &group.Name, &group.OwnerID,
		&group.ProfilePicURL, &group.CreatedAt, &group.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get group by id: %w", err)
	}
	return &group, nil
}

func (r *postgresGroupRepository) GetByHandle(ctx context.Context, handle string) (*domain.Group, error) {
	query := `SELECT id, handle, name, owner_id, profile_pic_url, created_at, updated_at
              FROM groups WHERE handle = $1`
	var group domain.Group
	err := r.db.QueryRow(ctx, query, handle).Scan(
		&group.ID, &group.Handle, &group.Name, &group.OwnerID,
		&group.ProfilePicURL, &group.CreatedAt, &group.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get group by handle: %w", err)
	}
	return &group, nil
}

func (r *postgresGroupRepository) Update(ctx context.Context, group *domain.Group) error {
	query := `UPDATE groups SET name = $1, owner_id = $2, profile_pic_url = $3, updated_at = NOW()
              WHERE id = $4`
	_, err := r.db.Exec(ctx, query, group.Name, group.OwnerID, group.ProfilePicURL, group.ID)
	if err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}
	return nil
}

func (r *postgresGroupRepository) FindMember(ctx context.Context, groupID, userID string) (*domain.GroupMember, error) {
	query := `SELECT group_id, user_id, role, created_at FROM group_members
              WHERE group_id = $1 AND user_id = $2`
	var member domain.GroupMember
	err := r.db.QueryRow(ctx, query, groupID, userID).Scan(
		&member.GroupID, &member.UserID, &member.Role, &member.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to find group member: %w", err)
	}
	return &member, nil
}

func (r *postgresGroupRepository) RemoveMember(ctx context.Context, groupID, userID string) error {
	query := `DELETE FROM group_members WHERE group_id = $1 AND user_id = $2`
	cmdTag, err := r.db.Exec(ctx, query, groupID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove group member: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (r *postgresGroupRepository) GetMembersWithUserDetails(ctx context.Context, groupID string) ([]*domain.GroupMember, error) {
	query := `SELECT gm.group_id, gm.user_id, gm.role, gm.created_at,
                     u.username, u.profile_pic_url
              FROM group_members gm
              JOIN users u ON gm.user_id = u.id
              WHERE gm.group_id = $1
              ORDER BY gm.created_at`
	rows, err := r.db.Query(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	defer rows.Close()

	var members []*domain.GroupMember
	for rows.Next() {
		var member domain.GroupMember
		member.User = &domain.User{}
		err := rows.Scan(
			&member.GroupID, &member.UserID, &member.Role, &member.CreatedAt,
			&member.User.Username, &member.User.ProfilePicURL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group member: %w", err)
		}
		member.User.ID = member.UserID
		members = append(members, &member)
	}
	return members, nil
}

func (r *postgresGroupRepository) GetGroupsByUserID(ctx context.Context, userID string) ([]*domain.Group, error) {
	query := `SELECT g.id, g.handle, g.name, g.owner_id, g.profile_pic_url, g.created_at, g.updated_at
              FROM groups g
              JOIN group_members gm ON g.id = gm.group_id
              WHERE gm.user_id = $1
              ORDER BY g.name`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups by user id: %w", err)
	}
	defer rows.Close()

	var groups []*domain.Group
	for rows.Next() {
		var group domain.Group
		err := rows.Scan(
			&group.ID, &group.Handle, &group.Name, &group.OwnerID,
			&group.ProfilePicURL, &group.CreatedAt, &group.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, &group)
	}
	return groups, nil
}

func (r *postgresGroupRepository) SearchByHandle(ctx context.Context, query string) ([]*domain.Group, error) {
	sqlQuery := `SELECT id, handle, name, owner_id, profile_pic_url, created_at, updated_at
                 FROM groups WHERE handle ILIKE $1 LIMIT 20`
	rows, err := r.db.Query(ctx, sqlQuery, "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to search groups: %w", err)
	}
	defer rows.Close()

	var groups []*domain.Group
	for rows.Next() {
		var group domain.Group
		err := rows.Scan(
			&group.ID, &group.Handle, &group.Name, &group.OwnerID,
			&group.ProfilePicURL, &group.CreatedAt, &group.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group during search: %w", err)
		}
		groups = append(groups, &group)
	}
	return groups, nil
}

func (r *postgresGroupRepository) GetOldestMember(ctx context.Context, groupID, excludeUserID string) (*domain.GroupMember, error) {
	query := `SELECT group_id, user_id, role, created_at FROM group_members
              WHERE group_id = $1 AND user_id != $2
              ORDER BY created_at ASC LIMIT 1`
	var member domain.GroupMember
	err := r.db.QueryRow(ctx, query, groupID, excludeUserID).Scan(
		&member.GroupID, &member.UserID, &member.Role, &member.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get oldest group member: %w", err)
	}
	return &member, nil
}

