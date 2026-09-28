package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type nullValuesTestModel struct {
	Price float64  `json:"price"`
	Note  *string  `json:"note"`
	Tags  []string `json:"tags"`
	Name  string   `json:"name"`
}

func TestPopulateFromMap_JSONNull(t *testing.T) {
	noteVal := "initial"
	s := nullValuesTestModel{
		Price: 100.0,
		Note:  &noteVal,
		Tags:  []string{"tag1", "tag2"},
		Name:  "keep",
	}

	assert.NotPanics(t, func() {
		populateFromMap(&s, map[string]interface{}{
			"price": nil,
			"note":  nil,
			"tags":  nil,
			"name":  nil,
		})
	})

	assert.Nil(t, s.Note)
	assert.Nil(t, s.Tags)
	assert.Equal(t, 100.0, s.Price)
	assert.Equal(t, "keep", s.Name)

	// Normal value still works
	populateFromMap(&s, map[string]interface{}{"price": 12.5})
	assert.Equal(t, 12.5, s.Price)
}
