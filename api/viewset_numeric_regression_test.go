package api

import (
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetIntFieldRejectsUnsignedOverflow(t *testing.T) {
	for _, value := range []interface{}{uint64(math.MaxUint64), uint(math.MaxUint64), namedUnsigned(math.MaxUint64)} {
		var destination int64 = 7
		require.Error(t, setIntField(reflect.ValueOf(&destination).Elem(), value))
		require.Equal(t, int64(7), destination)
	}
}

type namedUnsigned uint64
