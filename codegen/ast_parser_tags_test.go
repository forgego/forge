package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildTagsFromGOFLAGS(t *testing.T) {
	assert.Empty(t, buildTagsFromGOFLAGS(""))
	assert.Equal(t, []string{"a", "b"}, buildTagsFromGOFLAGS("-mod=mod -tags=a,b"))
	assert.Equal(t, []string{"c"}, buildTagsFromGOFLAGS("--tags c"))
	assert.Equal(t, []string{"d"}, buildTagsFromGOFLAGS("-tags=a -tags=d"))
}
