package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Relations support documented cascade policies
func TestRelationsSupportDocumentedCascadePolicies(t *testing.T) {
	cascades := []CascadeType{
		CascadeCASCADE,
		CascadeSET_NULL,
		CascadePROTECT,
		CascadeSET_DEFAULT,
		CascadeDO_NOTHING,
	}

	expectedValues := map[CascadeType]string{
		CascadeCASCADE:     "CASCADE",
		CascadeSET_NULL:    "SET_NULL",
		CascadePROTECT:     "PROTECT",
		CascadeSET_DEFAULT: "SET_DEFAULT",
		CascadeDO_NOTHING:  "DO_NOTHING",
	}

	for _, cascade := range cascades {
		assert.Equal(t, expectedValues[cascade], string(cascade))

		rel := Relation{
			Name:     "parent",
			To:       "Category",
			Type:     RelationForeignKey,
			OnDelete: cascade,
			OnUpdate: cascade,
		}
		assert.Equal(t, cascade, rel.OnDelete)
		assert.Equal(t, cascade, rel.OnUpdate)
	}
}

// Many-to-many relations support through tables
func TestManyToManyRelationsSupportThroughTables(t *testing.T) {
	rel := Relation{
		Name:    "groups",
		To:      "Group",
		Type:    RelationManyToMany,
		Through: "membership_table",
	}

	assert.Equal(t, RelationManyToMany, rel.Type)
	assert.Equal(t, "Group", rel.To)
	assert.Equal(t, "membership_table", rel.Through)
}
