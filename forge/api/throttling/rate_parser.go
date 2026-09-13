package throttling

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parseRate parses a rate string like "100/hour" or "1000/day"
// Returns (limit, duration, error)
func parseRate(rate string) (int, time.Duration, error) {
	parts := strings.Split(rate, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid rate format: %s", rate)
	}

	limit, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("invalid limit: %s", parts[0])
	}

	period := strings.TrimSpace(strings.ToLower(parts[1]))
	var duration time.Duration

	switch {
	case strings.HasPrefix(period, "sec"):
		duration = time.Second
	case strings.HasPrefix(period, "min"):
		duration = time.Minute
	case strings.HasPrefix(period, "hour") || strings.HasPrefix(period, "hr"):
		duration = time.Hour
	case strings.HasPrefix(period, "day"):
		duration = 24 * time.Hour
	default:
		return 0, 0, fmt.Errorf("invalid period: %s", period)
	}

	return limit, duration, nil
}
