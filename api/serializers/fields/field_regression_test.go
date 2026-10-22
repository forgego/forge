package fields

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIntegerFieldRejectsFractionalAndOverflow(t *testing.T) {
	field := NewIntegerField("count")
	for _, value := range []interface{}{1.5, math.Inf(1), float64(1e20)} {
		_, err := field.ToInternalValue(value)
		require.Error(t, err)
	}
}

func TestWriteOnlyFieldValidatesSource(t *testing.T) {
	source := NewStringField("password")
	min := 4
	source.MinLength = &min
	field := NewWriteOnlyField("password", source)
	require.Error(t, field.Validate("a"))
}
