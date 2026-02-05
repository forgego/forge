package repository

import (
	"context"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/identity/models"
	"github.com/forgego/forge/identity/testutils"
	"github.com/forgego/forge/identity/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB creates a database for testing
func setupTestDB(t *testing.T) *db.DB {
	return testutils.SetupTestDB(t)
}

func TestUserRepository_Create(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	repo := NewUserRepository(testDB)
	ctx := context.Background()

	t.Run("creates user successfully", func(t *testing.T) {
		hashedPassword, err := utils.HashPassword("testpassword123")
		require.NoError(t, err)

		user := &models.User{
			Username:  "johndoe",
			Email:     "john@example.com",
			Password:  hashedPassword,
			FirstName: "John",
			LastName:  "Doe",
			IsActive:  true,
		}

		err = repo.Create(ctx, user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
		assert.Equal(t, "johndoe", user.Username)
		assert.Equal(t, "john@example.com", user.Email)
	})

	t.Run("fails with duplicate email", func(t *testing.T) {
		hashedPassword, _ := utils.HashPassword("testpassword123")

		user1 := &models.User{
			Username: "user1",
			Email:    "duplicate@example.com",
			Password: hashedPassword,
		}
		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		user2 := &models.User{
			Username: "user2",
			Email:    "duplicate@example.com", // Same email
			Password: hashedPassword,
		}
		err = repo.Create(ctx, user2)
		assert.Error(t, err)
		// Postgres duplicate key error check
		assert.Contains(t, err.Error(), "duplicate key value violates unique constraint")
	})

	t.Run("fails with duplicate username", func(t *testing.T) {
		hashedPassword, _ := utils.HashPassword("testpassword123")

		user1 := &models.User{
			Username: "duplicate_user",
			Email:    "user1@example.com",
			Password: hashedPassword,
		}
		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		user2 := &models.User{
			Username: "duplicate_user", // Same username
			Email:    "user2@example.com",
			Password: hashedPassword,
		}
		err = repo.Create(ctx, user2)
		assert.Error(t, err)
		// Postgres duplicate key error check
		assert.Contains(t, err.Error(), "duplicate key value violates unique constraint")
	})

	t.Run("normalizes email to lowercase", func(t *testing.T) {
		hashedPassword, _ := utils.HashPassword("testpassword123")

		user := &models.User{
			Username: "testuser",
			Email:    "Test@EXAMPLE.COM", // Mixed case
			Password: hashedPassword,
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Retrieve and verify email is normalized
		retrieved, err := repo.GetByEmail(ctx, "test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", retrieved.Email)
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	repo := NewUserRepository(testDB)
	ctx := context.Background()

	// Create test user
	hashedPassword, _ := utils.HashPassword("testpassword123")
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		IsActive: true,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("retrieves existing user", func(t *testing.T) {
		retrieved, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Username, retrieved.Username)
		assert.Equal(t, user.Email, retrieved.Email)
		assert.True(t, retrieved.IsActive)
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		_, err := repo.GetByID(ctx, 99999)
		assert.Error(t, err)
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	repo := NewUserRepository(testDB)
	ctx := context.Background()

	hashedPassword, _ := utils.HashPassword("testpassword123")
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("retrieves user by email", func(t *testing.T) {
		retrieved, err := repo.GetByEmail(ctx, "test@example.com")
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Email, retrieved.Email)
	})

	t.Run("retrieves user by email case-insensitive", func(t *testing.T) {
		retrieved, err := repo.GetByEmail(ctx, "TEST@EXAMPLE.COM")
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
	})

	t.Run("returns error for non-existent email", func(t *testing.T) {
		_, err := repo.GetByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
	})
}

func TestUserRepository_GetByUsername(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	repo := NewUserRepository(testDB)
	ctx := context.Background()

	hashedPassword, _ := utils.HashPassword("testpassword123")
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("retrieves user by username", func(t *testing.T) {
		retrieved, err := repo.GetByUsername(ctx, "testuser")
		require.NoError(t, err)
		assert.Equal(t, user.ID, retrieved.ID)
		assert.Equal(t, user.Username, retrieved.Username)
	})

	t.Run("returns error for non-existent username", func(t *testing.T) {
		_, err := repo.GetByUsername(ctx, "nonexistent")
		assert.Error(t, err)
	})
}

func TestUserRepository_Update(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	repo := NewUserRepository(testDB)
	ctx := context.Background()

	hashedPassword, _ := utils.HashPassword("testpassword123")
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		IsActive: true,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("updates user fields", func(t *testing.T) {
		user.FirstName = "Updated"
		user.LastName = "Name"
		user.IsActive = false

		err := repo.Update(ctx, user)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated", updated.FirstName)
		assert.Equal(t, "Name", updated.LastName)
		assert.False(t, updated.IsActive)
	})

	t.Run("fails to update with duplicate email", func(t *testing.T) {
		// Create another user
		user2 := &models.User{
			Username: "user2",
			Email:    "user2@example.com",
			Password: hashedPassword,
		}
		err := repo.Create(ctx, user2)
		require.NoError(t, err)

		// Try to update user2's email to user1's email
		user2.Email = user.Email
		err = repo.Update(ctx, user2)
		assert.Error(t, err)
	})
}

func TestUserRepository_List(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	repo := NewUserRepository(testDB)
	ctx := context.Background()

	hashedPassword, _ := utils.HashPassword("testpassword123")

	// Create multiple users
	users := []*models.User{
		{Username: "user1", Email: "user1@example.com", Password: hashedPassword, IsActive: true, IsStaff: true},
		{Username: "user2", Email: "user2@example.com", Password: hashedPassword, IsActive: true, IsStaff: false},
		{Username: "user3", Email: "user3@example.com", Password: hashedPassword, IsActive: false, IsStaff: false},
	}

	for _, u := range users {
		err := repo.Create(ctx, u)
		require.NoError(t, err)
	}

	t.Run("lists all users", func(t *testing.T) {
		filters := &UserFilters{Limit: 10}
		result, err := repo.List(ctx, filters)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 3)
	})

	t.Run("filters by is_active", func(t *testing.T) {
		active := true
		filters := &UserFilters{IsActive: &active, Limit: 10}
		result, err := repo.List(ctx, filters)
		require.NoError(t, err)
		for _, u := range result {
			assert.True(t, u.IsActive)
		}
	})

	t.Run("filters by is_staff", func(t *testing.T) {
		staff := true
		filters := &UserFilters{IsStaff: &staff, Limit: 10}
		result, err := repo.List(ctx, filters)
		require.NoError(t, err)
		for _, u := range result {
			assert.True(t, u.IsStaff)
		}
	})

	t.Run("filters by email", func(t *testing.T) {
		filters := &UserFilters{Email: "user1@example.com", Limit: 10}
		result, err := repo.List(ctx, filters)
		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "user1@example.com", result[0].Email)
	})

	t.Run("applies pagination", func(t *testing.T) {
		filters := &UserFilters{Limit: 2, Offset: 0}
		result, err := repo.List(ctx, filters)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(result), 2)
	})
}

func TestUserRepository_Count(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	repo := NewUserRepository(testDB)
	ctx := context.Background()

	hashedPassword, _ := utils.HashPassword("testpassword123")

	// Create test users
	for i := 0; i < 5; i++ {
		user := &models.User{
			Username: "user" + string(rune('0'+i)),
			Email:    "user" + string(rune('0'+i)) + "@example.com",
			Password: hashedPassword,
			IsActive: i%2 == 0, // Alternate active/inactive
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)
	}

	t.Run("counts all users", func(t *testing.T) {
		filters := &UserFilters{}
		count, err := repo.Count(ctx, filters)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(5))
	})

	t.Run("counts filtered users", func(t *testing.T) {
		active := true
		filters := &UserFilters{IsActive: &active}
		count, err := repo.Count(ctx, filters)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(3)) // At least 3 active users
	})
}

func TestUserRepository_Exists(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	repo := NewUserRepository(testDB)
	ctx := context.Background()

	hashedPassword, _ := utils.HashPassword("testpassword123")
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("returns true for existing email", func(t *testing.T) {
		exists, err := repo.Exists(ctx, "test@example.com")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("returns false for non-existent email", func(t *testing.T) {
		exists, err := repo.Exists(ctx, "nonexistent@example.com")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	testDB := setupTestDB(t)
	defer testDB.Close()

	repo := NewUserRepository(testDB)
	ctx := context.Background()

	hashedPassword, _ := utils.HashPassword("testpassword123")
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
	}
	err := repo.Create(ctx, user)
	require.NoError(t, err)

	t.Run("soft deletes user", func(t *testing.T) {
		err := repo.Delete(ctx, user.ID)
		require.NoError(t, err)

		// User should not be retrievable via GetByID (filters out deleted)
		_, err = repo.GetByID(ctx, user.ID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")

		// But should still exist in database (verify via direct query)
		var deletedAt interface{}
		err = testDB.QueryRow("SELECT deleted_at FROM users WHERE id = $1", user.ID).Scan(&deletedAt)
		require.NoError(t, err)
		assert.NotNil(t, deletedAt)
	})

	t.Run("returns error for non-existent user", func(t *testing.T) {
		err := repo.Delete(ctx, 99999)
		assert.Error(t, err)
	})
}
