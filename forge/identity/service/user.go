package service

import (
	"context"
	"fmt"
	"time"

	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/repository"
	"github.com/forgego/forge/identity/utils"
)

// Predefined errors
var (
	ErrUserNotFound         = fmt.Errorf("user not found")
	ErrEmailExists          = fmt.Errorf("email already exists")
	ErrUsernameExists       = fmt.Errorf("username already exists")
	ErrInvalidEmail         = fmt.Errorf("invalid email address")
	ErrInvalidUsername      = fmt.Errorf("invalid username")
	ErrUserInactive         = fmt.Errorf("user account is inactive")
	ErrUserLocked           = fmt.Errorf("user account is locked")
	ErrEmailNotVerified     = fmt.Errorf("email not verified")
	ErrInvalidToken         = fmt.Errorf("invalid verification token")
	ErrTokenExpired         = fmt.Errorf("verification token has expired")
	ErrTokenAlreadyUsed     = fmt.Errorf("verification token has already been used")
	ErrEmailAlreadyVerified = fmt.Errorf("email is already verified")
)

// userService implements UserService interface
type userService struct {
	userRepo    repository.UserRepository
	tokenRepo   repository.TokenRepository
	emailSender EmailSender
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepository, tokenRepo repository.TokenRepository, emailSender EmailSender) UserService {
	return &userService{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		emailSender: emailSender,
	}
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) (*models.User, error) {
	// Validate request
	if err := validateCreateUserRequest(req); err != nil {
		return nil, err
	}

	// Check if email exists
	exists, err := s.userRepo.Exists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Check if username exists
	exists, err = s.userRepo.ExistsUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if exists {
		return nil, ErrUsernameExists
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	now := time.Now()
	user := &models.User{
		Username:      req.Username,
		Email:         req.Email,
		Password:      hashedPassword,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		IsActive:      req.IsActive,
		IsStaff:       req.IsStaff,
		IsSuperuser:   false, // Only set via CreateSuperuser
		IsLocked:      false,
		EmailVerified: false,
		DateJoined:    now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Save to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *userService) UpdateUser(ctx context.Context, id int64, req *UpdateUserRequest) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Update fields if provided
	if req.Email != nil {
		// Check if new email already exists
		if *req.Email != user.Email {
			exists, err := s.userRepo.Exists(ctx, *req.Email)
			if err != nil {
				return nil, fmt.Errorf("failed to check email: %w", err)
			}
			if exists {
				return nil, ErrEmailExists
			}
		}
		user.Email = *req.Email
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}

	if req.LastName != nil {
		user.LastName = *req.LastName
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if req.IsStaff != nil {
		user.IsStaff = *req.IsStaff
	}

	if req.IsLocked != nil {
		user.IsLocked = *req.IsLocked
		if *req.IsLocked {
			now := time.Now()
			user.LockedAt = &now
		} else {
			user.LockedAt = nil
			user.LockedReason = ""
		}
	}

	if req.IsSuperuser != nil {
		user.IsSuperuser = *req.IsSuperuser
	}

	// Update in database
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// DeleteUser deletes a user (soft delete)
func (s *userService) DeleteUser(ctx context.Context, id int64) error {
	// Check if user exists
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	// Delete user
	if err := s.userRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// GetUser retrieves a user by ID
func (s *userService) GetUser(ctx context.Context, id int64) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// ListUsers retrieves users with filters
func (s *userService) ListUsers(ctx context.Context, filters *UserFilters) ([]*models.User, int64, error) {
	// Convert service filters to repository filters
	repoFilters := &repository.UserFilters{
		Email:         filters.Email,
		Username:      filters.Username,
		IsActive:      filters.IsActive,
		IsStaff:       filters.IsStaff,
		IsSuperuser:   filters.IsSuperuser,
		IsLocked:      filters.IsLocked,
		EmailVerified: filters.EmailVerified,
		Search:        filters.Search,
		Limit:         filters.Limit,
		Offset:        filters.Offset,
		OrderBy:       filters.OrderBy,
	}

	// Get users
	users, err := s.userRepo.List(ctx, repoFilters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// Get count
	count, err := s.userRepo.Count(ctx, repoFilters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	return users, count, nil
}

// Register registers a new user (public endpoint)
func (s *userService) Register(ctx context.Context, req *RegisterRequest) (*models.User, error) {
	// Create user request
	createReq := &CreateUserRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		IsActive: true, // Active by default, but email verification may be required
		IsStaff:  false,
	}

	// Create user
	user, err := s.CreateUser(ctx, createReq)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// VerifyEmail verifies a user's email address
func (s *userService) VerifyEmail(ctx context.Context, token string) error {
	// 1. Find token in database
	verificationToken, err := s.tokenRepo.GetEmailVerificationToken(ctx, token)
	if err != nil {
		return ErrInvalidToken
	}

	// 2. Check if token is expired
	if verificationToken.IsExpired() {
		return ErrTokenExpired
	}

	// 3. Check if token was already used
	if verificationToken.IsUsed() {
		return ErrTokenAlreadyUsed
	}

	// 4. Get the user
	user, err := s.userRepo.GetByID(ctx, verificationToken.UserID)
	if err != nil {
		return ErrUserNotFound
	}

	// 5. Check if email is already verified
	if user.EmailVerified {
		return ErrEmailAlreadyVerified
	}

	// 6. Mark user's email as verified
	now := time.Now()
	user.EmailVerified = true
	user.EmailVerifiedAt = &now

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// 7. Mark token as used
	if err := s.tokenRepo.MarkEmailVerificationTokenUsed(ctx, verificationToken.ID); err != nil {
		// Log this error but don't fail - the email is verified
		// In production, you might want to handle this differently
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	return nil
}

// ResendVerificationEmail resends verification email
func (s *userService) ResendVerificationEmail(ctx context.Context, email string) error {
	// 1. Find user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not for security
		return nil
	}

	// 2. Check if email is already verified
	if user.EmailVerified {
		return ErrEmailAlreadyVerified
	}

	// 3. Create a new verification token
	verificationToken, err := s.CreateEmailVerificationToken(ctx, user.ID, user.Email)
	if err != nil {
		return fmt.Errorf("failed to create verification token: %w", err)
	}

	// 4. Send verification email
	if err := s.emailSender.SendVerificationEmail(ctx, user.Email, verificationToken.Token); err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	return nil
}

// CreateEmailVerificationToken creates a new email verification token
func (s *userService) CreateEmailVerificationToken(ctx context.Context, userID int64, email string) (*models.EmailVerificationToken, error) {
	// Generate secure random token
	token, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Hash the token for storage (for security)
	hashedToken, err := utils.HashPassword(token)
	if err != nil {
		return nil, fmt.Errorf("failed to hash token: %w", err)
	}

	// Create verification token record
	verificationToken := &models.EmailVerificationToken{
		UserID:    userID,
		Token:     hashedToken,
		Email:     email,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Save to database
	if err := s.tokenRepo.CreateEmailVerificationToken(ctx, verificationToken); err != nil {
		return nil, fmt.Errorf("failed to create verification token: %w", err)
	}

	// Return the token with the unhashed value for sending
	verificationToken.Token = token
	return verificationToken, nil
}

// validateCreateUserRequest validates a create user request
func validateCreateUserRequest(req *CreateUserRequest) error {
	if req.Username == "" {
		return ErrInvalidUsername
	}

	if req.Email == "" {
		return ErrInvalidEmail
	}

	if req.Password == "" {
		return fmt.Errorf("password is required")
	}

	// Basic email validation (more thorough validation should be done in serializer)
	if len(req.Email) > 254 {
		return ErrInvalidEmail
	}

	// Basic username validation
	if len(req.Username) > 150 {
		return ErrInvalidUsername
	}

	return nil
}

