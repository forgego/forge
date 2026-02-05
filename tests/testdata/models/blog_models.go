package testdata

import (
	"database/sql"
	"time"

	"github.com/forgego/forge/schema"
)

// Blog represents a main blog model
type Blog struct {
	schema.BaseSchema
	ID          int64          `db:"id"`
	Title       string         `db:"title"`
	Description sql.NullString `db:"description"`
	AuthorID    int64          `db:"author_id"` // FK to Author
	IsActive    bool           `db:"is_active"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

// Post represents blog posts
type Post struct {
	schema.BaseSchema
	ID          int64          `db:"id"`
	BlogID      int64          `db:"blog_id"`   // FK to Blog
	AuthorID    int64          `db:"author_id"` // FK to Author
	Title       string         `db:"title"`
	Content     string         `db:"content"`
	Excerpt     sql.NullString `db:"excerpt"`
	Status      string         `db:"status"` // draft, published, archived
	PublishedAt sql.NullTime   `db:"published_at"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

// Comment represents comments on posts with nested replies support
type Comment struct {
	schema.BaseSchema
	ID         int64         `db:"id"`
	PostID     int64         `db:"post_id"`   // FK to Post
	AuthorID   int64         `db:"author_id"` // FK to Author
	ParentID   sql.NullInt64 `db:"parent_id"` // Self-referential FK for replies
	Content    string        `db:"content"`
	IsApproved bool          `db:"is_approved"`
	CreatedAt  time.Time     `db:"created_at"`
	UpdatedAt  time.Time     `db:"updated_at"`
}

// Tag represents tags for posts (many-to-many)
type Tag struct {
	schema.BaseSchema
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Slug      string    `db:"slug"` // Unique
	CreatedAt time.Time `db:"created_at"`
}

// PostTag represents the many-to-many relationship between Post and Tag
type PostTag struct {
	schema.BaseSchema
	ID     int64 `db:"id"`
	PostID int64 `db:"post_id"` // FK to Post
	TagID  int64 `db:"tag_id"`  // FK to Tag
}

// Author represents blog authors
type Author struct {
	schema.BaseSchema
	ID          int64          `db:"id"`
	Username    string         `db:"username"` // Unique
	Email       string         `db:"email"`    // Unique
	DisplayName sql.NullString `db:"display_name"`
	Bio         sql.NullString `db:"bio"`
	AvatarURL   sql.NullString `db:"avatar_url"`
	IsActive    bool           `db:"is_active"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

// Category represents post categories (hierarchical)
type Category struct {
	schema.BaseSchema
	ID          int64          `db:"id"`
	Name        string         `db:"name"`
	Slug        string         `db:"slug"` // Unique
	Description sql.NullString `db:"description"`
	ParentID    sql.NullInt64  `db:"parent_id"` // Self-referential FK for hierarchy
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

// PostCategory represents the many-to-many relationship between Post and Category
type PostCategory struct {
	schema.BaseSchema
	ID         int64 `db:"id"`
	PostID     int64 `db:"post_id"`     // FK to Post
	CategoryID int64 `db:"category_id"` // FK to Category
}
