package db

import (
	"context"
	"testing"
)

func TestNamedSavepointRollsBackOnlyLaterWrites(t *testing.T) {
	database := setupTransactionTestDB(t)
	err := database.WithTx(context.Background(), func(tx *Tx) error {
		if _, err := tx.Exec("INSERT INTO test_items (id,name) VALUES (1,'kept')"); err != nil {
			return err
		}
		sp, err := tx.CreateSavepoint("before_second")
		if err != nil {
			return err
		}
		if _, err := tx.Exec("INSERT INTO test_items (id,name) VALUES (2,'removed')"); err != nil {
			return err
		}
		if err := sp.RollbackToSavepoint(); err != nil {
			return err
		}
		return sp.ReleaseSavepoint()
	})
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := database.QueryRow("SELECT count(*) FROM test_items").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("rows after partial rollback = %d", count)
	}
}
