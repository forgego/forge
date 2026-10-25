package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type checkMissingGetQueryset struct {
	unusedListOperations
}

func (*checkMissingGetQueryset) Create(context.Context, *checkedModel) error { return nil }
func (*checkMissingGetQueryset) Update(context.Context, *checkedModel) error { return nil }
func (*checkMissingGetQueryset) Delete(context.Context, *checkedModel) error { return nil }

type checkGetWrongReturnArity struct{ checkWritableQueryset }

func (*checkGetWrongReturnArity) Get(context.Context, int64) *checkedModel { return nil }

func TestViewSetConfiguration_MissingGet(t *testing.T) {
	vs := newCheckViewSet(&checkMissingGetQueryset{})
	message := registerPanic("widgets", vs)
	require.NotEmpty(t, message, "Register must panic")
	assert.Contains(t, message, `api: resource "widgets": `)
	assert.Contains(t, message, "Get")
}

func TestViewSetConfiguration_GetWrongArgType(t *testing.T) {
	vs := newCheckViewSet(&checkBadGet{})
	message := registerPanic("widgets", vs)
	require.NotEmpty(t, message, "Register must panic")
	assert.Contains(t, message, `api: resource "widgets": `)
	assert.Contains(t, message, "Get")
}

func TestViewSetConfiguration_GetWrongReturnArity(t *testing.T) {
	vs := newCheckViewSet(&checkGetWrongReturnArity{})
	message := registerPanic("widgets", vs)
	require.NotEmpty(t, message, "Register must panic")
	assert.Contains(t, message, `api: resource "widgets": `)
	assert.Contains(t, message, "Get")
}

func TestViewSetConfiguration_ValidReadWriteRegisters(t *testing.T) {
	vs := newCheckViewSet(&checkWritableQueryset{})
	message := registerPanic("widgets", vs)
	assert.Empty(t, message, "Register must not panic for valid read-write viewset")

	router := NewRouter("/api")
	require.NotPanics(t, func() {
		router.Register("widgets", vs)
	})
	assert.Equal(t, vs, router.routes["widgets"])
}
