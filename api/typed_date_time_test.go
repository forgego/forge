package api

import (
	"testing"
	"time"
)

func TestDateAndTimestampResponsesFollowSchemaLayouts(t *testing.T) {
	instant := time.Date(2026, 10, 21, 14, 15, 16, 0, time.UTC)
	model := schemaInputModel{StartsAt: instant, DateOnly: instant, OccurredAt: instant}
	result := modelToMap(model)
	for name, want := range map[string]string{"starts_at": "2026-10-21T14:15:16Z", "date_only": "2026-10-21", "occurred_at": "2026-10-21T14:15:16Z"} {
		if got := result[name]; got != want {
			t.Fatalf("%s = %v, want %q", name, got, want)
		}
	}
}
