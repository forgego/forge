package orm

import (
	"context"
	"testing"

	forgeerrors "github.com/forgego/forge/errors"
)

func TestManagerGetMissingRecordReturnsTypedNotFound(t *testing.T) {
	database := setupRelationTestDB(t)
	t.Cleanup(func() { _ = database.Close() })
	manager, err := NewManagerWithDB[TestCompany]("companies", database)
	if err != nil {
		t.Fatal(err)
	}
	got, err := manager.Get(context.Background(), 987654)
	if got != nil || !forgeerrors.IsNotFound(err) {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestQuerySetFirstReturnsNilForEmptyResult(t *testing.T) {
	database := setupRelationTestDB(t)
	t.Cleanup(func() { _ = database.Close() })
	qs, err := NewQuerySet[TestCompany]("companies")
	if err != nil {
		t.Fatal(err)
	}
	got, err := qs.SetDB(database).First(context.Background())
	if err != nil || got != nil {
		t.Fatalf("got=%v err=%v", got, err)
	}
}

func TestQuerySetUpdateReturnsAffectedRowCount(t *testing.T) {
	database := setupRelationTestDB(t)
	t.Cleanup(func() { _ = database.Close() })
	_, err := database.Exec(`INSERT INTO companies (id, name) VALUES (1, 'one'), (2, 'two'), (3, 'three')`)
	if err != nil {
		t.Fatal(err)
	}
	qs, err := NewQuerySet[TestCompany]("companies")
	if err != nil {
		t.Fatal(err)
	}
	count, err := qs.SetDB(database).Filter(F("id").In(int64(1), int64(2))).Update(context.Background(), UpdateMap{"name": "updated"})
	if err != nil || count != 2 {
		t.Fatalf("count=%d err=%v", count, err)
	}
	var updated int
	if err := database.QueryRow(`SELECT count(*) FROM companies WHERE name = 'updated'`).Scan(&updated); err != nil || updated != 2 {
		t.Fatalf("updated=%d err=%v", updated, err)
	}
}

func TestQuerySetDeleteReturnsDeletedRowCount(t *testing.T) {
	database := setupRelationTestDB(t)
	t.Cleanup(func() { _ = database.Close() })
	_, err := database.Exec(`INSERT INTO companies (id, name) VALUES (1, 'one'), (2, 'two'), (3, 'three')`)
	if err != nil {
		t.Fatal(err)
	}
	qs, err := NewQuerySet[TestCompany]("companies")
	if err != nil {
		t.Fatal(err)
	}
	count, err := qs.SetDB(database).Filter(F("id").In(int64(1), int64(2))).Delete(context.Background())
	if err != nil || count != 2 {
		t.Fatalf("count=%d err=%v", count, err)
	}
	var remaining int
	if err := database.QueryRow(`SELECT count(*) FROM companies`).Scan(&remaining); err != nil || remaining != 1 {
		t.Fatalf("remaining=%d err=%v", remaining, err)
	}
}
