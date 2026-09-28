package filters

import (
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type managerFilterModel struct {
	schema.BaseSchema
	ID     int64  `json:"id" db:"id"`
	Secret string `json:"secret" db:"secret_hash"`
}

func (managerFilterModel) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary(), schema.AutoIncrement()),
		{Name: "secret", DBColumn: "secret_hash", Type: schema.TypeString, Editable: true, Serialize: false},
	}
}

type managerBackedFilterQueryset struct {
	*orm.Manager[managerFilterModel]
	filtered orm.Expression
	ordered  []string
}

func (q *managerBackedFilterQueryset) Filter(expr orm.Expression) interface{} {
	q.filtered = expr
	return q
}

func (q *managerBackedFilterQueryset) OrderBy(fields ...string) interface{} {
	q.ordered = append(q.ordered, fields...)
	return q
}

func newManagerBackedFilterQueryset(t *testing.T) *managerBackedFilterQueryset {
	t.Helper()
	manager, err := orm.NewManager[managerFilterModel]("manager_filter_models")
	require.NoError(t, err)
	return &managerBackedFilterQueryset{Manager: manager}
}

func TestSearchFilter_ManagerSchemaExcludesNonSerializableColumn(t *testing.T) {
	queryset := newManagerBackedFilterQueryset(t)
	request := httptest.NewRequest("GET", "/test/?search=candidate", nil)
	NewSearchFilter([]string{"secret"}).FilterQueryset(request, queryset)
	assert.Nil(t, queryset.filtered)
}

func TestOrderingFilter_ManagerSchemaExcludesNonSerializableColumn(t *testing.T) {
	queryset := newManagerBackedFilterQueryset(t)
	request := httptest.NewRequest("GET", "/test/?ordering=secret", nil)
	NewOrderingFilter([]string{"secret"}).FilterQueryset(request, queryset)
	assert.Empty(t, queryset.ordered)
}
