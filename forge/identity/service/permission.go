package service

import (
	"context"
	"fmt"

	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
)

// permissionService implements PermissionService interface
type permissionService struct {
	permissionRepo repository.PermissionRepository
	userRepo       repository.UserRepository
}

// NewPermissionService creates a new permission service
func NewPermissionService(
	permissionRepo repository.PermissionRepository,
	userRepo repository.UserRepository,
) PermissionService {
	return &permissionService{
		permissionRepo: permissionRepo,
		userRepo:       userRepo,
	}
}

// CheckPermission checks if a user has a permission
func (s *permissionService) CheckPermission(ctx context.Context, userID int64, permission string) (bool, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, ErrUserNotFound
	}

	// Superusers have all permissions
	if user.IsSuperuser && user.IsActive {
		return true, nil
	}

	// Check permission
	hasPermission, err := s.permissionRepo.UserHasPermission(ctx, userID, permission)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return hasPermission, nil
}

// CheckPermissions checks if a user has all specified permissions
func (s *permissionService) CheckPermissions(ctx context.Context, userID int64, permissions []string) (bool, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, ErrUserNotFound
	}

	// Superusers have all permissions
	if user.IsSuperuser && user.IsActive {
		return true, nil
	}

	// Check all permissions
	for _, permission := range permissions {
		hasPermission, err := s.permissionRepo.UserHasPermission(ctx, userID, permission)
		if err != nil {
			return false, fmt.Errorf("failed to check permission: %w", err)
		}
		if !hasPermission {
			return false, nil
		}
	}

	return true, nil
}

// CheckAnyPermission checks if a user has any of the specified permissions
func (s *permissionService) CheckAnyPermission(ctx context.Context, userID int64, permissions []string) (bool, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return false, ErrUserNotFound
	}

	// Superusers have all permissions
	if user.IsSuperuser && user.IsActive {
		return true, nil
	}

	// Check any permission
	for _, permission := range permissions {
		hasPermission, err := s.permissionRepo.UserHasPermission(ctx, userID, permission)
		if err != nil {
			return false, fmt.Errorf("failed to check permission: %w", err)
		}
		if hasPermission {
			return true, nil
		}
	}

	return false, nil
}

// GetUserPermissions retrieves all permissions for a user
func (s *permissionService) GetUserPermissions(ctx context.Context, userID int64) ([]*models.Permission, error) {
	permissions, err := s.permissionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return permissions, nil
}

// AssignPermission assigns a permission to a user
func (s *permissionService) AssignPermission(ctx context.Context, userID int64, permission string) error {
	// Get permission
	perm, err := s.permissionRepo.GetByCodename(ctx, permission)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	// Assign to user
	if err := s.permissionRepo.AssignToUser(ctx, userID, perm.ID); err != nil {
		return fmt.Errorf("failed to assign permission: %w", err)
	}

	return nil
}

// RemovePermission removes a permission from a user
func (s *permissionService) RemovePermission(ctx context.Context, userID int64, permission string) error {
	// Get permission
	perm, err := s.permissionRepo.GetByCodename(ctx, permission)
	if err != nil {
		return fmt.Errorf("permission not found: %w", err)
	}

	// Remove from user
	if err := s.permissionRepo.RemoveFromUser(ctx, userID, perm.ID); err != nil {
		return fmt.Errorf("failed to remove permission: %w", err)
	}

	return nil
}
