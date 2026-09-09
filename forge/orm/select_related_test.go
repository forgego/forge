package orm

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/internal/testutils"
	"github.com/forgego/forge/schema"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type User struct {
	schema.BaseSchema
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}

func (User) Meta() schema.Meta {
	return schema.Meta{
		TableName: "users",
	}
}

func (User) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name", schema.Required()),
		schema.StringField("email", schema.Unique()),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

type Post struct {
	schema.BaseSchema
	ID        int64     `db:"id"`
	Title     string    `db:"title"`
	Content   string    `db:"content"`
	UserID    int64     `db:"user_id"`
	User      *User     `db:"user"` // Relation
	CreatedAt time.Time `db:"created_at"`
}

func (Post) Meta() schema.Meta {
	return schema.Meta{
		TableName: "posts",
	}
}

func (Post) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title", schema.Required()),
		schema.StringField("content"),
		schema.Int64Field("user_id"),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (Post) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("User", "User", schema.OnDelete(schema.CascadeCASCADE)),
	}
}

func TestSelectRelated_Integration(t *testing.T) {
	sqlDB := testutils.SetupTestDB(t)
	// Create tables manually for now as migration system is separate
	_, err := sqlDB.Exec(`
		DROP TABLE IF EXISTS posts CASCADE;
		DROP TABLE IF EXISTS users CASCADE;
		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			created_at TIMESTAMP
		);
		CREATE TABLE posts (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			content TEXT,
			user_id INTEGER,
			created_at TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id)
		);
	`)
	require.NoError(t, err)
	defer sqlDB.Close()

	database := &db.DB{DB: sqlDB, Driver: "postgres"}

	// Insert data
	var userID int64
	err = database.QueryRow(`INSERT INTO users (name, email, created_at) VALUES ($1, $2, $3) RETURNING id`, "John Doe", "john@example.com", time.Now()).Scan(&userID)
	require.NoError(t, err)

	_, err = database.Exec(`INSERT INTO posts (title, content, user_id, created_at) VALUES ($1, $2, $3, $4)`, "My Post", "Hello World", userID, time.Now())
	require.NoError(t, err)

	// Test SelectRelated
	// Ensure User schema is registered for relation resolution
	_, err = GetModelSchema[User]()
	require.NoError(t, err)

	qs, err := NewQuerySet[Post]("posts")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	// 1. Without SelectRelated, User should be nil
	posts, err := qs.All(context.Background())
	require.NoError(t, err)
	require.Len(t, posts, 1)
	assert.Nil(t, posts[0].User)
	assert.Equal(t, userID, posts[0].UserID)

	// 2. With SelectRelated, User should be populated
	postsRelated, err := qs.SelectRelated("User").All(context.Background())
	require.NoError(t, err)
	require.Len(t, postsRelated, 1)
	require.NotNil(t, postsRelated[0].User, "User should be populated with SelectRelated")
	assert.Equal(t, "John Doe", postsRelated[0].User.Name)
	assert.Equal(t, userID, postsRelated[0].User.ID)
}

func TestSelectRelated_SQLite(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "select_related_test.sqlite")

	database, err := db.NewDB(dbPath)
	if err != nil {
		t.Skipf("skipping sqlite test: %v", err)
	}
	defer database.Close()

	_, err = database.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			created_at TIMESTAMP
		);
		CREATE TABLE posts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT,
			user_id INTEGER,
			created_at TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id)
		);
	`)
	require.NoError(t, err)

	now := time.Now().Truncate(time.Second)
	res, err := database.Exec(`INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)`, "Jane Doe", "jane@example.com", now)
	require.NoError(t, err)
	userID, err := res.LastInsertId()
	require.NoError(t, err)

	_, err = database.Exec(`INSERT INTO posts (title, content, user_id, created_at) VALUES (?, ?, ?, ?)`, "SQLite Post", "SQLite Content", userID, now)
	require.NoError(t, err)

	_, err = GetModelSchema[User]()
	require.NoError(t, err)

	qs, err := NewQuerySet[Post]("posts")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	// 1. Without SelectRelated, User should be nil
	posts, err := qs.All(context.Background())
	require.NoError(t, err)
	require.Len(t, posts, 1)
	assert.Nil(t, posts[0].User)
	assert.Equal(t, userID, posts[0].UserID)

	// 2. With SelectRelated("User"), User should be populated
	postsRelated, err := qs.SelectRelated("User").All(context.Background())
	require.NoError(t, err)
	require.Len(t, postsRelated, 1)
	require.NotNil(t, postsRelated[0].User, "User should be populated with SelectRelated")
	assert.Equal(t, "Jane Doe", postsRelated[0].User.Name)
	assert.Equal(t, userID, postsRelated[0].User.ID)

	// 3. With lowercase SelectRelated("user"), User should also be populated (case-insensitivity)
	postsLower, err := qs.SelectRelated("user").All(context.Background())
	require.NoError(t, err)
	require.Len(t, postsLower, 1)
	require.NotNil(t, postsLower[0].User, "User should be populated with case-insensitive SelectRelated")
	assert.Equal(t, "Jane Doe", postsLower[0].User.Name)
	assert.Equal(t, userID, postsLower[0].User.ID)
}
