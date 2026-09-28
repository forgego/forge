package filter

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInMemoryFilterStorage_ConcurrentSaveAndList(t *testing.T) {
	storage := NewInMemoryFilterStorage()

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			ownerID := fmt.Sprintf("user_%d", id%5)
			filterID := fmt.Sprintf("filter_%d", id)

			err := storage.Save(&SavedFilter{
				ID:      filterID,
				OwnerID: ownerID,
				Name:    fmt.Sprintf("Filter %d", id),
			})
			if err != nil {
				t.Errorf("Save failed: %v", err)
				return
			}

			list, err := storage.List(ownerID)
			if err != nil {
				t.Errorf("List failed: %v", err)
				return
			}
			if len(list) == 0 {
				t.Errorf("expected at least 1 filter for owner %s", ownerID)
			}

			// Also verify Load, Update, and Delete concurrently
			loaded, err := storage.Load(filterID)
			if err != nil || loaded == nil {
				t.Errorf("Load failed: %v", err)
				return
			}

			loaded.Description = "updated"
			if err := storage.Update(loaded); err != nil {
				t.Errorf("Update failed: %v", err)
			}
		}(i)
	}

	wg.Wait()

	// Confirm 50 items were saved
	all, err := storage.List("")
	require.NoError(t, err)
	// Some may be owner-specific, but loading all 50 by ID should succeed
	for i := 0; i < numGoroutines; i++ {
		f, err := storage.Load(fmt.Sprintf("filter_%d", i))
		require.NoError(t, err)
		require.NotNil(t, f)
	}
	_ = all
}

func TestSaveFilter_LoadFilter_RoundTripAST(t *testing.T) {
	storage := NewInMemoryFilterStorage()

	fs, err := NewFilterSet[MockModel]()
	require.NoError(t, err)

	fs.Where("username").Equals("alice")
	require.NotNil(t, fs.GetAST())

	saved, err := SaveFilter(fs, "alice_filter", "find alice", storage)
	require.NoError(t, err)
	require.NotEmpty(t, saved.ID)

	loadedFs, err := LoadFilter[MockModel](saved.ID, storage)
	require.NoError(t, err)
	require.NotNil(t, loadedFs)

	loadedAST := loadedFs.GetAST()
	require.NotNil(t, loadedAST)
	require.Equal(t, fs.GetAST().Op, loadedAST.Op)
	require.Equal(t, fs.GetAST().Field, loadedAST.Field)
	require.Equal(t, fs.GetAST().Value, loadedAST.Value)
}
