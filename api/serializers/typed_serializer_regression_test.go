package serializers

import (
	"fmt"
	"math"
	"testing"

	"github.com/forgego/forge/orm"
	"github.com/forgego/forge/schema"
	"github.com/stretchr/testify/require"
)

type rejectForbidden struct{}

func (rejectForbidden) Validate(value interface{}) error {
	if value == "bad" {
		return fmt.Errorf("forbidden value")
	}
	return nil
}

type constrainedRecord struct {
	schema.BaseSchema
	ID       int64   `json:"id" db:"id"`
	Counter  int64   `json:"counter" db:"counter"`
	Computed string  `json:"computed" db:"computed"`
	Name     string  `json:"name" db:"name"`
	Status   string  `json:"status" db:"status"`
	Score    int64   `json:"score" db:"score"`
	Ratio    float32 `json:"ratio" db:"ratio"`
}

func (constrainedRecord) Fields() []schema.Field {
	return []schema.Field{
		schema.Int64Field("id", schema.Primary()),
		schema.Int64Field("counter", schema.AutoIncrement()),
		schema.StringField("computed", schema.GeneratedColumn("name", true)),
		schema.StringField("name", schema.MaxLength(3), schema.MinLength(2), schema.Validators(rejectForbidden{})),
		schema.StringField("status", schema.ChoicesFromPairsOpts("ok", "OK")),
		schema.Int64Field("score", schema.MinValue(2), schema.MaxValue(5)),
		schema.Float32Field("ratio"),
	}
}

func TestFromSchemaReadOnlyAndConstraints(t *testing.T) {
	s := FromSchema[constrainedRecord](constrainedRecord{})
	model := constrainedRecord{}
	require.NoError(t, s.Deserialize(map[string]interface{}{"id": int64(1), "counter": int64(2), "computed": "x"}, &model))
	require.Zero(t, model.ID)
	require.Zero(t, model.Counter)
	require.Empty(t, model.Computed)
	for _, test := range []struct {
		name  string
		value interface{}
	}{
		{"name", "long"}, {"name", "x"}, {"name", "bad"},
		{"status", "bad"}, {"score", int64(1)}, {"score", int64(6)},
	} {
		t.Run(fmt.Sprintf("%s_%v", test.name, test.value), func(t *testing.T) {
			model := constrainedRecord{}
			require.Error(t, s.Deserialize(map[string]interface{}{test.name: test.value}, &model))
			require.Contains(t, s.Errors(), test.name)
		})
	}
}

func TestFromSchemaTypedNil(t *testing.T) {
	var model *constrainedRecord
	s := FromSchema[constrainedRecord](model)
	_, err := s.Serialize(&constrainedRecord{})
	require.Error(t, err)
}

func TestTypedRejectsPrecisionLoss(t *testing.T) {
	s := FromSchema[constrainedRecord](constrainedRecord{})
	model := constrainedRecord{}
	require.Error(t, s.Deserialize(map[string]interface{}{"ratio": float64(1.1)}, &model))
	require.Contains(t, s.Errors(), "ratio")
	require.NoError(t, s.Deserialize(map[string]interface{}{"ratio": float64(1.5)}, &model))
	require.Equal(t, float32(1.5), model.Ratio)
	require.Error(t, s.Deserialize(map[string]interface{}{"ratio": math.MaxFloat64}, &model))
}

type serializerRecord struct {
	schema.BaseSchema
	Name   string `json:"name" db:"name"`
	Age    int64  `json:"age" db:"age"`
	Secret string `json:"secret" db:"secret"`
}

func (serializerRecord) Fields() []schema.Field {
	return []schema.Field{
		schema.StringField("name", schema.Required()),
		schema.Int64Field("age"),
		schema.StringField("secret", schema.WriteOnly()),
	}
}

func TestFromSchemaConcreteFieldsAndVisibility(t *testing.T) {
	s := FromSchema[serializerRecord](serializerRecord{})
	output, err := s.Serialize(&serializerRecord{Name: "Ada", Age: 42, Secret: "hidden"})
	require.NoError(t, err)
	require.Equal(t, map[string]interface{}{"name": "Ada", "age": int64(42)}, output)
	require.Error(t, s.Deserialize(map[string]interface{}{}, &serializerRecord{}))
	require.Contains(t, s.Errors(), "name")
}

func TestTypedRequiredNullAndWrongType(t *testing.T) {
	s := NewTypedSerializer[serializerRecord]().AddField("name", Field[serializerRecord](orm.NewField[string]("Name", "")).Required())
	for _, value := range []interface{}{nil, 12} {
		model := serializerRecord{Name: "previous"}
		require.Error(t, s.Deserialize(map[string]interface{}{"name": value}, &model))
		require.Contains(t, s.Errors(), "name")
		require.Equal(t, "previous", model.Name)
	}
}

func TestTypedRequiredDefaultAndNilDestination(t *testing.T) {
	s := NewTypedSerializer[serializerRecord]().AddField("name", Field[serializerRecord](orm.NewField[string]("Name", "")).Required().Default("fallback"))
	model := serializerRecord{}
	require.NoError(t, s.Deserialize(map[string]interface{}{}, &model))
	require.Equal(t, "fallback", model.Name)
	require.Error(t, s.Deserialize(map[string]interface{}{}, nil))
	_, err := s.Serialize(nil)
	require.Error(t, err)
}

func TestTypedWriteOnlyAndInvalidField(t *testing.T) {
	s := NewTypedSerializer[serializerRecord]().AddField("secret", Field[serializerRecord](orm.NewField[string]("Secret", "")).WriteOnly())
	output, err := s.Serialize(&serializerRecord{Secret: "hidden"})
	require.NoError(t, err)
	require.NotContains(t, output, "secret")
	s.AddField("missing", Field[serializerRecord](orm.NewField[string]("Missing", "")))
	require.Error(t, s.Deserialize(map[string]interface{}{"missing": "value"}, &serializerRecord{}))
}

func TestTypedRejectsLossyNumericCoercion(t *testing.T) {
	s := NewTypedSerializer[serializerRecord]().AddField("age", Field[serializerRecord](orm.NewField[int64]("Age", "")))
	for _, value := range []interface{}{1.5, uint64(^uint64(0)), "12"} {
		model := serializerRecord{Age: 7}
		require.Error(t, s.Deserialize(map[string]interface{}{"age": value}, &model))
		require.Equal(t, int64(7), model.Age)
	}
}
