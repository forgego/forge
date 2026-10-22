package repository

import "time"

// Identity's TIMESTAMP columns have no timezone. Store UTC wall time so
// database round trips preserve the instant used by expiry and lockout checks.
// Existing non-UTC rows need an explicit deployment-specific correction;
// their original timezone cannot be recovered from the stored timestamp.
func utcTimestamp(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}
