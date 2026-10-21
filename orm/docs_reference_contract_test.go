package orm

import (
	"context"
	"errors"
	"testing"

	"github.com/forgego/forge/db"
	"github.com/forgego/forge/schema"
)

type referenceModel struct{ ID int64 }

type referenceHookModel struct {
	schema.BaseSchema
	ID                int64  `db:"id"`
	Name              string `db:"name"`
	Value             int64  `db:"value"`
	AfterCreateCalled bool   `db:"-"`
}

func (referenceHookModel) Meta() schema.Meta { return schema.Meta{TableName: "tx_test_items"} }
func (referenceHookModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		schema.StringField("name"), schema.Int64Field("value"),
	}
}
func (m *referenceHookModel) AfterCreate(context.Context) error {
	m.AfterCreateCalled = true
	return nil
}

func TestAfterCreateRunsBeforeTransactionCommit_A183(t *testing.T) {
	database, base := setupTxTestDB(t)
	_ = base
	mgr, err := NewManagerWithDB[referenceHookModel]("tx_test_items", database)
	if err != nil {
		t.Fatal(err)
	}
	rollback := errors.New("rollback")
	model := &referenceHookModel{Name: "uncommitted", Value: 1}
	err = database.WithTx(context.Background(), func(tx *db.Tx) error {
		if err := mgr.WithTx(tx).Create(context.Background(), model); err != nil {
			return err
		}
		if !model.AfterCreateCalled {
			t.Fatal("AfterCreate had not run inside transaction")
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("transaction error = %v", err)
	}
	count, err := mgr.Count(context.Background())
	if err != nil || count != 0 {
		t.Fatalf("rollback count = %d, err = %v", count, err)
	}
}

// These assignments compile only while the documented public signatures hold.
func TestManagerReferenceSignatures(t *testing.T) {
	t.Run("A-191", func(t *testing.T) {
		var method func(*Manager[referenceModel], context.Context, *referenceModel) error = (*Manager[referenceModel]).Create
		_ = method
	})
	t.Run("A-192", func(t *testing.T) {
		var method func(*Manager[referenceModel], context.Context, int64) (*referenceModel, error) = (*Manager[referenceModel]).Get
		_ = method
	})
	t.Run("A-195", func(t *testing.T) {
		var method func(*Manager[referenceModel], context.Context, *referenceModel) error = (*Manager[referenceModel]).Delete
		_ = method
	})
	t.Run("A-196", func(t *testing.T) {
		var method func(*Manager[referenceModel], context.Context) ([]*referenceModel, error) = (*Manager[referenceModel]).All
		_ = method
	})
	t.Run("A-197", func(t *testing.T) {
		var method func(*Manager[referenceModel], Expression) (QuerySet[referenceModel], error) = (*Manager[referenceModel]).Filter
		_ = method
	})
	t.Run("A-202", func(t *testing.T) {
		var method func(*Manager[referenceModel], context.Context, []*referenceModel) error = (*Manager[referenceModel]).BulkCreate
		_ = method
	})
}

func TestQuerySetReferenceSignatures(t *testing.T) {
	t.Run("A-205", func(t *testing.T) {
		var method func(QuerySet[referenceModel], Expression) QuerySet[referenceModel] = QuerySet[referenceModel].Filter
		_ = method
	})
	t.Run("A-206", func(t *testing.T) {
		var method func(QuerySet[referenceModel], Expression) QuerySet[referenceModel] = QuerySet[referenceModel].Exclude
		_ = method
	})
	t.Run("A-215", func(t *testing.T) {
		var method func(QuerySet[referenceModel], ...Aggregate) QuerySet[referenceModel] = QuerySet[referenceModel].Aggregate
		var values func(context.Context, QuerySet[referenceModel], ...Aggregate) (map[string]any, error) = AggregateValues[referenceModel]
		_, _ = method, values
	})
	t.Run("A-216", func(t *testing.T) {
		var method func(QuerySet[referenceModel], ...AnnotationExpr) QuerySet[referenceModel] = QuerySet[referenceModel].Annotate
		_ = method
	})
	t.Run("A-218", func(t *testing.T) {
		var method func(QuerySet[referenceModel], context.Context) (*referenceModel, error) = QuerySet[referenceModel].Get
		_ = method
	})
}
