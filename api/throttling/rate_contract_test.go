package throttling

import (
	"net/http/httptest"
	"testing"
)

func TestInvalidRatesFailThrottleChecks(t *testing.T) {
	for _, rate := range []string{"garbage", "ten/hour", "10/fortnight"} {
		if _, _, err := parseRate(rate); err == nil {
			t.Fatalf("rate %q accepted", rate)
		}
	}
	throttle := NewAnonRateThrottle("garbage")
	_, _, err := throttle.AllowRequest(httptest.NewRequest("GET", "/", nil), nil)
	if err == nil {
		t.Fatal("invalid rate did not fail the throttle check")
	}
}
