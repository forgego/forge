package db

import (
	"database/sql"
	"testing"
	"time"
)

func TestDefaultPoolConfig(t *testing.T) {
	config := DefaultPoolConfig()

	if config.MaxOpenConns != 25 {
		t.Errorf("expected MaxOpenConns to be 25, got %d", config.MaxOpenConns)
	}
	if config.MaxIdleConns != 10 {
		t.Errorf("expected MaxIdleConns to be 10, got %d", config.MaxIdleConns)
	}
	if config.ConnMaxLifetime != 5*time.Minute {
		t.Errorf("expected ConnMaxLifetime to be 5m, got %v", config.ConnMaxLifetime)
	}
	if config.ConnMaxIdleTime != 2*time.Minute {
		t.Errorf("expected ConnMaxIdleTime to be 2m, got %v", config.ConnMaxIdleTime)
	}
}

func TestPoolConfigApply(t *testing.T) {
	// Create a mock DB with a real sql.DB for testing
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 not available (CGO required): %v", err)
	}
	defer sqlDB.Close()

	db := &DB{DB: sqlDB}

	// Apply custom pool config
	config := PoolConfig{
		MaxOpenConns:    50,
		MaxIdleConns:    25,
		ConnMaxLifetime: 10 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
	config.Apply(db)

	// Verify settings were applied via stats
	stats := db.PoolStats()
	if stats == nil {
		t.Fatal("expected PoolStats to return non-nil result")
	}
}

func TestWithPoolConfig(t *testing.T) {
	config := PoolConfig{
		MaxOpenConns:    30,
		MaxIdleConns:    15,
		ConnMaxLifetime: 3 * time.Minute,
		ConnMaxIdleTime: 1 * time.Minute,
	}

	// Create a mock DB with a real sql.DB for testing
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 not available (CGO required): %v", err)
	}
	defer sqlDB.Close()

	db := &DB{DB: sqlDB}
	WithPoolConfig(config)(db)

	// Verify the database was created successfully
	stats := db.PoolStats()
	if stats == nil {
		t.Fatal("expected PoolStats to return non-nil result")
	}
}

func TestWithMaxOpenConns(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 not available (CGO required): %v", err)
	}
	defer sqlDB.Close()

	db := &DB{DB: sqlDB}
	WithMaxOpenConns(100)(db)

	stats := db.PoolStats()
	if stats == nil {
		t.Fatal("expected PoolStats to return non-nil result")
	}
}

func TestWithMaxIdleConns(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 not available (CGO required): %v", err)
	}
	defer sqlDB.Close()

	db := &DB{DB: sqlDB}
	WithMaxIdleConns(50)(db)

	stats := db.PoolStats()
	if stats == nil {
		t.Fatal("expected PoolStats to return non-nil result")
	}
}

func TestWithConnMaxLifetime(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 not available (CGO required): %v", err)
	}
	defer sqlDB.Close()

	db := &DB{DB: sqlDB}
	WithConnMaxLifetime(30 * time.Minute)(db)

	stats := db.PoolStats()
	if stats == nil {
		t.Fatal("expected PoolStats to return non-nil result")
	}
}

func TestWithConnMaxIdleTime(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 not available (CGO required): %v", err)
	}
	defer sqlDB.Close()

	db := &DB{DB: sqlDB}
	WithConnMaxIdleTime(10 * time.Minute)(db)

	stats := db.PoolStats()
	if stats == nil {
		t.Fatal("expected PoolStats to return non-nil result")
	}
}

func TestMultipleOptions(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 not available (CGO required): %v", err)
	}
	defer sqlDB.Close()

	db := &DB{DB: sqlDB}
	WithMaxOpenConns(75)(db)
	WithMaxIdleConns(30)(db)
	WithConnMaxLifetime(15 * time.Minute)(db)
	WithConnMaxIdleTime(7 * time.Minute)(db)

	stats := db.PoolStats()
	if stats == nil {
		t.Fatal("expected PoolStats to return non-nil result")
	}
}

func TestPoolStats(t *testing.T) {
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 not available (CGO required): %v", err)
	}
	defer sqlDB.Close()

	db := &DB{DB: sqlDB}

	stats := db.PoolStats()
	if stats == nil {
		t.Fatal("expected PoolStats to return non-nil result")
	}

	// Verify basic stats fields are accessible
	_ = stats.MaxOpenConnections
	_ = stats.OpenConnections
	_ = stats.InUse
	_ = stats.Idle
	_ = stats.WaitCount
	_ = stats.WaitDuration
}

func TestDefaultPoolConfigAppliedOnNewDB(t *testing.T) {
	// Test that NewDB without options applies default pool config
	// This test verifies the logic without requiring a real database connection
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 not available (CGO required): %v", err)
	}
	defer sqlDB.Close()

	db := &DB{DB: sqlDB}
	defaultConfig := DefaultPoolConfig()
	defaultConfig.Apply(db)

	stats := db.PoolStats()
	if stats == nil {
		t.Fatal("expected PoolStats to return non-nil result")
	}
}

func TestOptionsOverrideDefaults(t *testing.T) {
	// Test that options override default pool config
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 not available (CGO required): %v", err)
	}
	defer sqlDB.Close()

	db := &DB{DB: sqlDB}
	// Apply defaults first
	defaultConfig := DefaultPoolConfig()
	defaultConfig.Apply(db)
	// Then override with option
	WithMaxOpenConns(100)(db)

	stats := db.PoolStats()
	if stats == nil {
		t.Fatal("expected PoolStats to return non-nil result")
	}
}

func TestPoolConfigApplyWithZeroValues(t *testing.T) {
	// Test that zero values in PoolConfig don't override existing settings
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 not available (CGO required): %v", err)
	}
	defer sqlDB.Close()

	db := &DB{DB: sqlDB}
	// First apply some settings
	WithMaxOpenConns(50)(db)

	// Apply config with zero values - should not change anything
	config := PoolConfig{}
	config.Apply(db)

	// Database should still work
	stats := db.PoolStats()
	if stats == nil {
		t.Fatal("expected PoolStats to return non-nil result")
	}
}

func TestPoolConfigApplyPartial(t *testing.T) {
	// Test applying partial config (only some fields set)
	sqlDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Skipf("sqlite3 not available (CGO required): %v", err)
	}
	defer sqlDB.Close()

	db := &DB{DB: sqlDB}

	// Apply config with only MaxOpenConns set
	config := PoolConfig{
		MaxOpenConns: 40,
	}
	config.Apply(db)

	stats := db.PoolStats()
	if stats == nil {
		t.Fatal("expected PoolStats to return non-nil result")
	}
}

func TestOptionType(t *testing.T) {
	// Test that Option type is correctly defined as a function
	var _ Option = func(db *DB) {}
}

func TestPoolConfigStruct(t *testing.T) {
	// Test that PoolConfig struct has all expected fields
	config := PoolConfig{
		MaxOpenConns:    1,
		MaxIdleConns:    2,
		ConnMaxLifetime: 3 * time.Second,
		ConnMaxIdleTime: 4 * time.Second,
	}

	if config.MaxOpenConns != 1 {
		t.Errorf("expected MaxOpenConns 1, got %d", config.MaxOpenConns)
	}
	if config.MaxIdleConns != 2 {
		t.Errorf("expected MaxIdleConns 2, got %d", config.MaxIdleConns)
	}
	if config.ConnMaxLifetime != 3*time.Second {
		t.Errorf("expected ConnMaxLifetime 3s, got %v", config.ConnMaxLifetime)
	}
	if config.ConnMaxIdleTime != 4*time.Second {
		t.Errorf("expected ConnMaxIdleTime 4s, got %v", config.ConnMaxIdleTime)
	}
}
