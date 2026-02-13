package orm

import (
	"context"
	"testing"
	"time"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/db/dialect"
	"github.com/forgego/forge/internal/testutils"
	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Tag struct {
	schema.BaseSchema
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

func (Tag) Meta() schema.Meta {
	return schema.Meta{TableName: "tags"}
}

func (Tag) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"),
	}
}

type Article struct {
	schema.BaseSchema
	ID        int64     `db:"id"`
	Title     string    `db:"title"`
	UserID    int64     `db:"user_id"`
	User      *User     `db:"user"` // Forward FK
	Tags      []Tag     `db:"tags"` // M2M
	CreatedAt time.Time `db:"created_at"`
}

func (Article) Meta() schema.Meta {
	return schema.Meta{TableName: "articles"}
}

func (Article) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("title"),
		schema.Int64Field("user_id"),
		schema.TimeField("created_at", schema.AutoNowAdd()),
	}
}

func (Article) Relations() []schema.Relation {
	return []schema.Relation{
		schema.ForeignKeyField("User", "User"),
		schema.ManyToManyField("Tags", "Tag", schema.Through("article_tags")),
	}
}

func TestPrefetchRelated_Integration(t *testing.T) {
	sqlDB := testutils.SetupTestDB(t)
	// Create tables
	_, err := sqlDB.Exec(`
		DROP TABLE IF EXISTS article_tags CASCADE;
		DROP TABLE IF EXISTS articles CASCADE;
		DROP TABLE IF EXISTS tags CASCADE;
		DROP TABLE IF EXISTS posts CASCADE; 
		DROP TABLE IF EXISTS users CASCADE;

		CREATE TABLE users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			created_at TIMESTAMP
		);
		CREATE TABLE tags (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL
		);
		CREATE TABLE articles (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			user_id INTEGER,
			created_at TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id)
		);
		CREATE TABLE article_tags (
			article_id INTEGER,
			tag_id INTEGER,
			PRIMARY KEY(article_id, tag_id),
			FOREIGN KEY(article_id) REFERENCES articles(id),
			FOREIGN KEY(tag_id) REFERENCES tags(id)
		);
	`)
	require.NoError(t, err)
	defer sqlDB.Close()

	database := &db.DB{DB: sqlDB, Driver: "postgres"}
	// Set the dialect manually since we're creating DB directly
	// In production, NewDB() would set this automatically
	database.SetDialect(dialect.NewPostgreSQLDialect())

	// Insert data
	var userID int64
	err = database.QueryRow(`INSERT INTO users (name, email, created_at) VALUES ($1, $2, $3) RETURNING id`, "Jane Doe", "jane@example.com", time.Now()).Scan(&userID)
	require.NoError(t, err)

	var tag1ID, tag2ID int64
	err = database.QueryRow(`INSERT INTO tags (name) VALUES ($1) RETURNING id`, "Go").Scan(&tag1ID)
	require.NoError(t, err)
	err = database.QueryRow(`INSERT INTO tags (name) VALUES ($1) RETURNING id`, "Forge").Scan(&tag2ID)
	require.NoError(t, err)

	var articleID int64
	err = database.QueryRow(`INSERT INTO articles (title, user_id, created_at) VALUES ($1, $2, $3) RETURNING id`, "Intro to Forge", userID, time.Now()).Scan(&articleID)
	require.NoError(t, err)

	_, err = database.Exec(`INSERT INTO article_tags (article_id, tag_id) VALUES ($1, $2), ($1, $3)`, articleID, tag1ID, tag2ID)
	require.NoError(t, err)

	// Ensure schemas are registered
	GetModelSchema[User]()
	GetModelSchema[Tag]()
	GetModelSchema[Article]()

	// 1. Test PrefetchRelated FK (User)
	qs, err := NewQuerySet[Article]("articles")
	require.NoError(t, err)
	qs = qs.SetDB(database)

	articles, err := qs.PrefetchRelated("User").All(context.Background())
	require.NoError(t, err)
	require.Len(t, articles, 1)
	assert.NotNil(t, articles[0].User)
	if articles[0].User != nil {
		assert.Equal(t, "Jane Doe", articles[0].User.Name)
	}

	// 2. Test PrefetchRelated M2M (Tags)
	articlesTags, err := qs.PrefetchRelated("Tags").All(context.Background())
	require.NoError(t, err)
	require.Len(t, articlesTags, 1)
	assert.Len(t, articlesTags[0].Tags, 2)
	
	// Verify tags
	tagNames := make(map[string]bool)
	for _, tag := range articlesTags[0].Tags {
		tagNames[tag.Name] = true
	}
	assert.True(t, tagNames["Go"])
	assert.True(t, tagNames["Forge"])
}
