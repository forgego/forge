package execute

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaselineReport_AllVerified(t *testing.T) {
	// Nil report
	var nilReport *BaselineReport
	assert.True(t, nilReport.AllVerified())

	// Empty report
	emptyReport := &BaselineReport{Entries: []BaselineEntry{}}
	assert.True(t, emptyReport.AllVerified())

	// All verified
	allVerified := &BaselineReport{
		Entries: []BaselineEntry{
			{Version: 1, Status: BaselineStatusVerified},
			{Version: 2, Status: BaselineStatusVerified},
			{Version: 3, Status: BaselineStatusVerified},
		},
	}
	assert.True(t, allVerified.AllVerified())

	// Contains unverified
	withUnverified := &BaselineReport{
		Entries: []BaselineEntry{
			{Version: 1, Status: BaselineStatusVerified},
			{Version: 2, Status: BaselineStatusUnverified},
		},
	}
	assert.False(t, withUnverified.AllVerified())

	// Contains mismatched
	withMismatched := &BaselineReport{
		Entries: []BaselineEntry{
			{Version: 1, Status: BaselineStatusVerified},
			{Version: 2, Status: BaselineStatusMismatched},
		},
	}
	assert.False(t, withMismatched.AllVerified())

	// Contains missing
	withMissing := &BaselineReport{
		Entries: []BaselineEntry{
			{Version: 1, Status: BaselineStatusVerified},
			{Version: 2, Status: BaselineStatusMissing},
		},
	}
	assert.False(t, withMissing.AllVerified())
}

func TestBaselineReport_HasBlocking(t *testing.T) {
	// Nil report
	var nilReport *BaselineReport
	assert.False(t, nilReport.HasBlocking())

	// Empty report
	emptyReport := &BaselineReport{Entries: []BaselineEntry{}}
	assert.False(t, emptyReport.HasBlocking())

	// All verified
	allVerified := &BaselineReport{
		Entries: []BaselineEntry{
			{Version: 1, Status: BaselineStatusVerified},
			{Version: 2, Status: BaselineStatusVerified},
		},
	}
	assert.False(t, allVerified.HasBlocking())

	// Unverified only (does not block)
	unverifiedOnly := &BaselineReport{
		Entries: []BaselineEntry{
			{Version: 1, Status: BaselineStatusUnverified},
			{Version: 2, Status: BaselineStatusUnverified},
		},
	}
	assert.False(t, unverifiedOnly.HasBlocking())

	// Verified and unverified (does not block)
	verifiedAndUnverified := &BaselineReport{
		Entries: []BaselineEntry{
			{Version: 1, Status: BaselineStatusVerified},
			{Version: 2, Status: BaselineStatusUnverified},
		},
	}
	assert.False(t, verifiedAndUnverified.HasBlocking())

	// Mismatched (blocks)
	mismatched := &BaselineReport{
		Entries: []BaselineEntry{
			{Version: 1, Status: BaselineStatusVerified},
			{Version: 2, Status: BaselineStatusMismatched},
		},
	}
	assert.True(t, mismatched.HasBlocking())

	// Missing (blocks)
	missing := &BaselineReport{
		Entries: []BaselineEntry{
			{Version: 1, Status: BaselineStatusMissing},
			{Version: 2, Status: BaselineStatusVerified},
		},
	}
	assert.True(t, missing.HasBlocking())
}

func TestBaselineReport_Counts(t *testing.T) {
	// Nil report
	var nilReport *BaselineReport
	nilCounts := nilReport.Counts()
	assert.Equal(t, 0, nilCounts[BaselineStatusVerified])
	assert.Equal(t, 0, nilCounts[BaselineStatusMismatched])
	assert.Equal(t, 0, nilCounts[BaselineStatusMissing])
	assert.Equal(t, 0, nilCounts[BaselineStatusUnverified])

	// Empty report
	emptyReport := &BaselineReport{Entries: []BaselineEntry{}}
	emptyCounts := emptyReport.Counts()
	assert.Equal(t, 0, emptyCounts[BaselineStatusVerified])
	assert.Equal(t, 0, emptyCounts[BaselineStatusMismatched])
	assert.Equal(t, 0, emptyCounts[BaselineStatusMissing])
	assert.Equal(t, 0, emptyCounts[BaselineStatusUnverified])

	// Mixed report
	report := &BaselineReport{
		Entries: []BaselineEntry{
			{Version: 1, Status: BaselineStatusVerified},
			{Version: 2, Status: BaselineStatusVerified},
			{Version: 3, Status: BaselineStatusMismatched},
			{Version: 4, Status: BaselineStatusMissing},
			{Version: 5, Status: BaselineStatusUnverified},
			{Version: 6, Status: BaselineStatusUnverified},
		},
	}
	counts := report.Counts()
	assert.Equal(t, 2, counts[BaselineStatusVerified])
	assert.Equal(t, 1, counts[BaselineStatusMismatched])
	assert.Equal(t, 1, counts[BaselineStatusMissing])
	assert.Equal(t, 2, counts[BaselineStatusUnverified])
}
