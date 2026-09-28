package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseFilterValue_NumbersBeforeBooleans pins how list query parameters
// are typed. "1" and "0" used to become booleans, so ?project_id=1 compared a
// bigint column with a boolean, which PostgreSQL rejects (a 500 response).
func TestParseFilterValue_NumbersBeforeBooleans(t *testing.T) {
	for raw, want := range map[string]interface{}{
		"1":     int64(1),
		"0":     int64(0),
		"42":    int64(42),
		"1.5":   1.5,
		"true":  true,
		"False": false,
		"t":     true,
		"abc":   "abc",
		"":      "",
	} {
		assert.Equal(t, want, parseFilterValue(raw), "parseFilterValue(%q)", raw)
	}
}
