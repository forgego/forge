package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/models"
)

// permissionRepository implements PermissionRepository interface
type permissionRepository struct {
	db *db.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(database *db.DB) PermissionRepository {
	return &permissionRepository{db: database}
}

// GetByID retrieves a permission by ID
func (r *permissionRepository) GetByID(ctx context.Context, id int64) (*models.Permission, error) {
	query := `
		SELECT id, name, codename, content_type, app_label
		FROM permissions
		WHERE id = $1
	`

	permission := &models.Permission{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&permission.ID,
		&permission.Name,
		&permission.Codename,
		&permission.ContentType,
		&permission.AppLabel,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return permission, nil
}

// GetByCodename retrieves a permission by codename
func (r *permissionRepository) GetByCodename(ctx context.Context, codename string) (*models.Permission, error) {
	query := `
		SELECT id, name, codename, content_type, app_label
		FROM permissions
		WHERE codename = $1
	`

	permission := &models.Permission{}
	err := r.db.QueryRowContext(ctx, query, codename).Scan(
		&permission.ID,
		&permission.Name,
		&permission.Codename,
		&permission.ContentType,
		&permission.AppLabel,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("permission not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return permission, nil
}

// GetByUserID retrieves all permissions for a user (direct + via groups)
func (r *permissionRepository) GetByUserID(ctx context.Context, userID int64) ([]*models.Permission, error) {
	query := `
		SELECT DISTINCT p.id, p.name, p.codename, p.content_type, p.app_label
		FROM permissions p
		LEFT JOIN user_permissions up ON p.id = up.permission_id
		LEFT JOIN group_permissions gp ON p.id = gp.permission_id
		LEFT JOIN user_groups ug ON gp.group_id = ug.group_id
		WHERE (up.user_id = $1 OR ug.user_id = $1)
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*models.Permission
	for rows.Next() {
		permission := &models.Permission{}
		err := rows.Scan(
			&permission.ID,
			&permission.Name,
			&permission.Codename,
			&permission.ContentType,
			&permission.AppLabel,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate permissions: %w", err)
	}

	return permissions, nil
}

// AssignToUser assigns a permission to a user
func (r *permissionRepository) AssignToUser(ctx context.Context, userID, permissionID int64) error {
	query := `
		INSERT INTO user_permissions (user_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, permission_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, userID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to assign permission to user: %w", err)
	}

	return nil
}

// RemoveFromUser removes a permission from a user
func (r *permissionRepository) RemoveFromUser(ctx context.Context, userID, permissionID int64) error {
	query := `DELETE FROM user_permissions WHERE user_id = $1 AND permission_id = $2`

	result, err := r.db.ExecContext(ctx, query, userID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove permission from user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission not assigned to user")
	}

	return nil
}

// UserHasPermission checks if a user has a specific permission
func (r *permissionRepository) UserHasPermission(ctx context.Context, userID int64, codename string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM permissions p
			LEFT JOIN user_permissions up ON p.id = up.permission_id
			LEFT JOIN group_permissions gp ON p.id = gp.permission_id
			LEFT JOIN user_groups ug ON gp.group_id = ug.group_id
			WHERE (up.user_id = $1 OR ug.user_id = $1)
			AND p.codename = $2
		)
	`

	var hasPermission bool
	err := r.db.QueryRowContext(ctx, query, userID, codename).Scan(&hasPermission)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return hasPermission, nil
}

// groupRepository implements GroupRepository interface
type groupRepository struct {
	db *db.DB
}

// NewGroupRepository creates a new group repository
func NewGroupRepository(database *db.DB) GroupRepository {
	return &groupRepository{db: database}
}

// Create creates a new group
func (r *groupRepository) Create(ctx context.Context, group *models.Group) error {
	query := `
		INSERT INTO groups (name, description)
		VALUES ($1, $2)
		RETURNING id
	`

	err := r.db.QueryRowContext(ctx, query, group.Name, group.Description).Scan(&group.ID)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	return nil
}

// GetByID retrieves a group by ID
func (r *groupRepository) GetByID(ctx context.Context, id int64) (*models.Group, error) {
	query := `
		SELECT id, name, description
		FROM groups
		WHERE id = $1
	`

	group := &models.Group{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&group.ID,
		&group.Name,
		&group.Description,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("group not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	return group, nil
}

// GetByName retrieves a group by name
func (r *groupRepository) GetByName(ctx context.Context, name string) (*models.Group, error) {
	query := `
		SELECT id, name, description
		FROM groups
		WHERE name = $1
	`

	group := &models.Group{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(
		&group.ID,
		&group.Name,
		&group.Description,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("group not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	return group, nil
}

// Update updates a group
func (r *groupRepository) Update(ctx context.Context, group *models.Group) error {
	query := `UPDATE groups SET name = $2, description = $3 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, group.ID, group.Name, group.Description)
	if err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("group not found")
	}

	return nil
}

// Delete deletes a group
func (r *groupRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM groups WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("group not found")
	}

	return nil
}

// List retrieves all groups
func (r *groupRepository) List(ctx context.Context) ([]*models.Group, error) {
	query := `SELECT id, name, description FROM groups ORDER BY name`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	defer rows.Close()

	var groups []*models.Group
	for rows.Next() {
		group := &models.Group{}
		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate groups: %w", err)
	}

	return groups, nil
}

// GetByUserID retrieves all groups for a user
func (r *groupRepository) GetByUserID(ctx context.Context, userID int64) ([]*models.Group, error) {
	query := `
		SELECT g.id, g.name, g.description
		FROM groups g
		INNER JOIN user_groups ug ON g.id = ug.group_id
		WHERE ug.user_id = $1
		ORDER BY g.name
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}
	defer rows.Close()

	var groups []*models.Group
	for rows.Next() {
		group := &models.Group{}
		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate groups: %w", err)
	}

	return groups, nil
}

// AddUser adds a user to a group
func (r *groupRepository) AddUser(ctx context.Context, groupID, userID int64) error {
	query := `
		INSERT INTO user_groups (user_id, group_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, group_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, userID, groupID)
	if err != nil {
		return fmt.Errorf("failed to add user to group: %w", err)
	}

	return nil
}

// RemoveUser removes a user from a group
func (r *groupRepository) RemoveUser(ctx context.Context, groupID, userID int64) error {
	query := `DELETE FROM user_groups WHERE user_id = $1 AND group_id = $2`

	result, err := r.db.ExecContext(ctx, query, userID, groupID)
	if err != nil {
		return fmt.Errorf("failed to remove user from group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not in group")
	}

	return nil
}

// AddPermission adds a permission to a group
func (r *groupRepository) AddPermission(ctx context.Context, groupID, permissionID int64) error {
	query := `
		INSERT INTO group_permissions (group_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (group_id, permission_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, groupID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to add permission to group: %w", err)
	}

	return nil
}

// RemovePermission removes a permission from a group
func (r *groupRepository) RemovePermission(ctx context.Context, groupID, permissionID int64) error {
	query := `DELETE FROM group_permissions WHERE group_id = $1 AND permission_id = $2`

	result, err := r.db.ExecContext(ctx, query, groupID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove permission from group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("permission not in group")
	}

	return nil
}
