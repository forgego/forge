package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type orderingTestModel struct {
	schema.BaseSchema
	ID        int64  `json:"id" db:"item_id"`
	Name      string `json:"displayName" db:"display_name"`
	metaOrder []string
}

func (orderingTestModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.DBColumn("item_id")),
		schema.StringField("name", schema.DBColumn("display_name")),
	}
}

func (m orderingTestModel) Meta() schema.Meta {
	return schema.Meta{TableName: "ordering_test_models", OrderBy: m.metaOrder}
}

type orderingTestQuerySet struct {
	fields      []interface{}
	filterCalls int
	filterExpr  orm.Expression
}

func (*orderingTestQuerySet) Count(context.Context) (int64, error) { return 0, nil }
func (qs *orderingTestQuerySet) Offset(int) interface{}            { return qs }
func (qs *orderingTestQuerySet) Limit(int) interface{}             { return qs }
func (qs *orderingTestQuerySet) Filter(expression interface{}) interface{} {
	qs.filterCalls++
	qs.filterExpr, _ = expression.(orm.Expression)
	return qs
}

func TestBaseViewSet_FilterAliasResolvesToDatabaseColumn(t *testing.T) {
	queryset := &orderingTestQuerySet{}
	viewSet := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, queryset, &orderingTestModel{})
	recorder := httptest.NewRecorder()
	viewSet.List(recorder, httptest.NewRequest(http.MethodGet, "/items/?displayName=chair", nil))

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.NotNil(t, queryset.filterExpr)
	sql, args, err := queryset.filterExpr.ToSQL(orm.NewSQLBuilder())
	require.NoError(t, err)
	assert.Equal(t, `"display_name" = $1`, sql)
	assert.Equal(t, []interface{}{"chair"}, args)
}
func (qs *orderingTestQuerySet) All(context.Context) ([]*orderingTestModel, error) {
	return nil, nil
}

func (qs *orderingTestQuerySet) OrderBy(fields ...interface{}) interface{} {
	qs.fields = append([]interface{}(nil), fields...)
	return qs
}

func TestBaseViewSet_ListAppliesStableDefaultOrdering(t *testing.T) {
	tests := []struct {
		name      string
		model     *orderingTestModel
		target    string
		wantField string
	}{
		{name: "model Meta ordering resolves schema name", model: &orderingTestModel{metaOrder: []string{"-name"}}, target: "/items/", wantField: "-display_name"},
		{name: "primary key fallback uses database column", model: &orderingTestModel{}, target: "/items/", wantField: "item_id"},
		{name: "explicit schema ordering resolves to column", model: &orderingTestModel{metaOrder: []string{"-name"}}, target: "/items/?ordering=name", wantField: "display_name"},
		{name: "explicit JSON ordering resolves to column", model: &orderingTestModel{}, target: "/items/?ordering=-displayName", wantField: "-display_name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs := &orderingTestQuerySet{}
			vs := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, qs, tt.model)
			recorder := httptest.NewRecorder()
			vs.List(recorder, httptest.NewRequest(http.MethodGet, tt.target, nil))

			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			assert.Equal(t, []interface{}{tt.wantField}, qs.fields)
		})
	}
}

type hiddenQueryModel struct {
	schema.BaseSchema
	ID       int64  `json:"id" db:"item_id"`
	Password string `json:"token" db:"secret_hash"`
}

func (hiddenQueryModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.DBColumn("item_id")),
		{Name: "secret", DBColumn: "secret_hash", Type: schema.TypeString, Editable: true, Serialize: false},
	}
}

func TestBaseViewSet_HiddenFieldAliasesCannotFilterOrOrder(t *testing.T) {
	aliases := []string{"secret", "secret_hash", "token", "Password"}
	for _, alias := range aliases {
		t.Run(alias, func(t *testing.T) {
			filterQS := &orderingTestQuerySet{}
			filterVS := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, filterQS, &hiddenQueryModel{})
			filterRec := httptest.NewRecorder()
			filterVS.List(filterRec, httptest.NewRequest(http.MethodGet, "/items/?"+alias+"=candidate", nil))
			require.Equal(t, http.StatusOK, filterRec.Code, filterRec.Body.String())
			assert.Zero(t, filterQS.filterCalls)

			orderingQS := &orderingTestQuerySet{}
			orderingVS := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, orderingQS, &hiddenQueryModel{})
			orderingRec := httptest.NewRecorder()
			orderingVS.List(orderingRec, httptest.NewRequest(http.MethodGet, "/items/?ordering="+alias, nil))
			require.Equal(t, http.StatusOK, orderingRec.Code, orderingRec.Body.String())
			assert.Empty(t, orderingQS.fields)
		})
	}
}
