package server

import (
	"testing"
	"time"
)

func TestKeyedLimiter_BurstAndIndependentKeys(t *testing.T) {
	limiter := NewKeyedLimiter(3, time.Minute)
	defer limiter.Close()
	for i := 0; i < 3; i++ {
		allowed, retryAfter := limiter.Reserve("client-a")
		if !allowed || retryAfter != 0 {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	allowed, retryAfter := limiter.Reserve("client-a")
	if allowed || retryAfter <= 0 {
		t.Fatal("exhausted key should be denied with a retry duration")
	}
	allowed, retryAfter = limiter.Reserve("client-b")
	if !allowed || retryAfter != 0 {
		t.Fatal("a different key should have an independent burst")
	}
}
