package testdata

import (
	"database/sql"
	"time"

	"github.com/forgego/forge/schema"
)

// User is a sample model for testing
type User struct {
	schema.BaseSchema
	ID        int64     `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Post is a sample model with FK relationship
type Post struct {
	schema.BaseSchema
	ID        int64          `db:"id"`
	Title     string         `db:"title"`
	Content   string         `db:"content"`
	UserID    int64          `db:"user_id"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
	Metadata  sql.NullString `db:"metadata"` // JSONB field
}

// Tag is a sample model for many-to-many testing
type Tag struct {
	schema.BaseSchema
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

// PostTag represents a many-to-many relationship
type PostTag struct {
	schema.BaseSchema
	ID     int64 `db:"id"`
	PostID int64 `db:"post_id"`
	TagID  int64 `db:"tag_id"`
}

// Category is for testing other features
type Category struct {
	schema.BaseSchema
	ID          int64          `db:"id"`
	Name        string         `db:"name"`
	Slug        string         `db:"slug"` // Unique field
	Description sql.NullString `db:"description"`
	CreatedAt   time.Time      `db:"created_at"`
}

// Product demonstrates a more complex model
type Product struct {
	schema.BaseSchema
	ID          int64          `db:"id"`
	SKU         string         `db:"sku"` // Unique
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Price       float64        `db:"price"`
	Stock       int32          `db:"stock"`
	CategoryID  int64          `db:"category_id"` // FK
	IsActive    bool           `db:"is_active"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}
