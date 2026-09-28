package execute

import (
	"testing"

	"github.com/forgego/forge/db/migrate/verify"
	"github.com/stretchr/testify/require"
)

func TestChecksumCompatibility(t *testing.T) {
	sql := "CREATE TABLE users (id SERIAL PRIMARY KEY, username VARCHAR(255));"

	execChecksum := CalculateChecksum(sql)
	verifyChecksum := verify.CalculateChecksum(sql)

	require.Equal(t, verifyChecksum, execChecksum)
	require.NoError(t, ValidateChecksum(sql, execChecksum))

	validator := NewChecksumValidator(t.TempDir())
	require.NotNil(t, validator)
}
