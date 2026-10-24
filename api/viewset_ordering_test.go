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

func TestBaseViewSet_RejectsUnknownAndUnsupportedFilters(t *testing.T) {
	for _, query := range []string{"?unknown=value", "?name__regex=value"} {
		t.Run(query, func(t *testing.T) {
			queryset := &orderingTestQuerySet{}
			viewSet := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, queryset, &orderingTestModel{})
			recorder := httptest.NewRecorder()
			viewSet.List(recorder, httptest.NewRequest(http.MethodGet, "/items/"+query, nil))

			assert.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
			assert.Contains(t, recorder.Body.String(), "filter")
			assert.Zero(t, queryset.filterCalls)
		})
	}
}

func TestBaseViewSet_AllowsListControlQueryParameters(t *testing.T) {
	queryset := &orderingTestQuerySet{}
	viewSet := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, queryset, &orderingTestModel{})
	recorder := httptest.NewRecorder()
	viewSet.List(recorder, httptest.NewRequest(http.MethodGet, "/items/?page=1&page_size=10&ordering=-name&search=desk&format=json", nil))

	assert.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
}

func TestBaseViewSet_EnforcesExplicitFilterableFields(t *testing.T) {
	queryset := &orderingTestQuerySet{}
	viewSet := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, queryset, &orderingTestModel{})
	viewSet.Filterable = []string{"name"}
	recorder := httptest.NewRecorder()
	viewSet.List(recorder, httptest.NewRequest(http.MethodGet, "/items/?id=1", nil))

	assert.Equal(t, http.StatusBadRequest, recorder.Code, recorder.Body.String())
	assert.Zero(t, queryset.filterCalls)
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
		name       string
		model      *orderingTestModel
		target     string
		wantFields []interface{}
	}{
		{name: "model Meta ordering resolves schema name", model: &orderingTestModel{metaOrder: []string{"-name"}}, target: "/items/", wantFields: []interface{}{"-display_name"}},
		{name: "primary key fallback uses database column", model: &orderingTestModel{}, target: "/items/", wantFields: []interface{}{"item_id"}},
		{name: "explicit schema ordering resolves to column", model: &orderingTestModel{metaOrder: []string{"-name"}}, target: "/items/?ordering=name", wantFields: []interface{}{"display_name", "item_id"}},
		{name: "explicit JSON ordering resolves to column", model: &orderingTestModel{}, target: "/items/?ordering=-displayName", wantFields: []interface{}{"-display_name", "item_id"}},
		{name: "empty ordering fields fall back to primary key", model: &orderingTestModel{}, target: "/items/?ordering=,%20,", wantFields: []interface{}{"item_id"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs := &orderingTestQuerySet{}
			vs := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, qs, tt.model)
			recorder := httptest.NewRecorder()
			vs.List(recorder, httptest.NewRequest(http.MethodGet, tt.target, nil))

			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			assert.Equal(t, tt.wantFields, qs.fields)
		})
	}
}

type hiddenPrimaryKeyOrderModel struct {
	schema.BaseSchema
	Key  int64  `json:"-" db:"private_key"`
	Name string `json:"name" db:"name"`
}

func (hiddenPrimaryKeyOrderModel) Fields() []schema.Field {
	return []schema.Field{
		{Name: "key", DBColumn: "private_key", Type: schema.TypeInt64, PrimaryKey: true, Serialize: false},
		schema.StringField("name"),
	}
}

type hiddenPrimaryKeyOrderQuerySet struct {
	fields []interface{}
}

func (*hiddenPrimaryKeyOrderQuerySet) Count(context.Context) (int64, error) { return 1, nil }
func (*hiddenPrimaryKeyOrderQuerySet) All(context.Context) ([]*hiddenPrimaryKeyOrderModel, error) {
	return []*hiddenPrimaryKeyOrderModel{{Key: 987654321, Name: "visible item"}}, nil
}
func (qs *hiddenPrimaryKeyOrderQuerySet) Offset(int) interface{} { return qs }
func (qs *hiddenPrimaryKeyOrderQuerySet) Limit(int) interface{}  { return qs }
func (qs *hiddenPrimaryKeyOrderQuerySet) OrderBy(fields ...interface{}) interface{} {
	qs.fields = append([]interface{}(nil), fields...)
	return qs
}

func TestBaseViewSet_HiddenPrimaryKeyOnlyControlsInternalDefaultOrdering(t *testing.T) {
	for _, order := range []string{"", "key", "-key", "private_key", "-private_key"} {
		t.Run(order, func(t *testing.T) {
			qs := &hiddenPrimaryKeyOrderQuerySet{}
			vs := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, qs, &hiddenPrimaryKeyOrderModel{})
			recorder := httptest.NewRecorder()
			vs.List(recorder, httptest.NewRequest(http.MethodGet, "/items/?ordering="+order, nil))
			require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
			assert.Equal(t, []interface{}{"private_key"}, qs.fields)
			assert.Contains(t, recorder.Body.String(), "visible item")
			assert.NotContains(t, recorder.Body.String(), "private_key")
			assert.NotContains(t, recorder.Body.String(), "987654321")
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
			require.Equal(t, http.StatusBadRequest, filterRec.Code, filterRec.Body.String())
			assert.Contains(t, filterRec.Body.String(), "invalid filter")
			assert.Zero(t, filterQS.filterCalls)

			orderingQS := &orderingTestQuerySet{}
			orderingVS := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, orderingQS, &hiddenQueryModel{})
			orderingRec := httptest.NewRecorder()
			orderingVS.List(orderingRec, httptest.NewRequest(http.MethodGet, "/items/?ordering="+alias, nil))
			require.Equal(t, http.StatusOK, orderingRec.Code, orderingRec.Body.String())
			assert.NotContains(t, orderingQS.fields, "secret_hash")
			assert.Equal(t, []interface{}{"item_id"}, orderingQS.fields)
		})
	}
}

type hiddenMetaOrderModel struct {
	schema.BaseSchema
	ID        int64  `json:"id" db:"item_id"`
	Password  string `json:"token" db:"secret_hash"`
	Name      string `json:"displayName" db:"display_name"`
	metaOrder []string
}

func (hiddenMetaOrderModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.DBColumn("item_id")),
		{Name: "secret", DBColumn: "secret_hash", Type: schema.TypeString, Editable: true, Serialize: false},
		schema.StringField("name", schema.DBColumn("display_name")),
	}
}

func (m hiddenMetaOrderModel) Meta() schema.Meta {
	return schema.Meta{TableName: "hidden_meta_order_models", OrderBy: m.metaOrder}
}

func TestBaseViewSet_HiddenOnlyMetaOrderByFallsBackToPrimaryKey(t *testing.T) {
	for _, order := range []string{"secret", "-secret", "secret_hash", "-secret_hash"} {
		t.Run(order, func(t *testing.T) {
			qs := &orderingTestQuerySet{}
			model := &hiddenMetaOrderModel{metaOrder: []string{order}}
			vs := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, qs, model)
			rec := httptest.NewRecorder()
			vs.List(rec, httptest.NewRequest(http.MethodGet, "/items/", nil))

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.NotContains(t, qs.fields, "secret_hash")
			assert.NotContains(t, qs.fields, "-secret_hash")
			assert.Equal(t, []interface{}{"item_id"}, qs.fields)
		})
	}
}

func TestBaseViewSet_ExplicitHiddenOrderingFallsBackToPrimaryKey(t *testing.T) {
	for _, order := range []string{"secret", "-secret", "secret_hash", "-secret_hash", "token", "-token", "Password", "-Password"} {
		t.Run(order, func(t *testing.T) {
			qs := &orderingTestQuerySet{}
			model := &hiddenMetaOrderModel{}
			vs := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, qs, model)
			rec := httptest.NewRecorder()
			vs.List(rec, httptest.NewRequest(http.MethodGet, "/items/?ordering="+order, nil))

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.NotContains(t, qs.fields, "secret_hash")
			assert.NotContains(t, qs.fields, "-secret_hash")
			assert.Equal(t, []interface{}{"item_id"}, qs.fields)
		})
	}
}

func TestBaseViewSet_MixedMetaOrderByKeepsVisibleFieldWithoutPrimaryKey(t *testing.T) {
	tests := []struct {
		name      string
		metaOrder []string
		wantField string
	}{
		{name: "hidden then visible ascending", metaOrder: []string{"secret", "name"}, wantField: "display_name"},
		{name: "hidden then visible descending", metaOrder: []string{"secret", "-name"}, wantField: "-display_name"},
		{name: "descending hidden then visible", metaOrder: []string{"-secret", "name"}, wantField: "display_name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qs := &orderingTestQuerySet{}
			model := &hiddenMetaOrderModel{metaOrder: tt.metaOrder}
			vs := NewBaseViewSet(func() Serializer { return NewBaseSerializer(nil) }, qs, model)
			rec := httptest.NewRecorder()
			vs.List(rec, httptest.NewRequest(http.MethodGet, "/items/", nil))

			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			assert.NotContains(t, qs.fields, "secret_hash")
			assert.NotContains(t, qs.fields, "-secret_hash")
			assert.NotContains(t, qs.fields, "item_id")
			assert.Equal(t, []interface{}{tt.wantField}, qs.fields)
		})
	}
}
