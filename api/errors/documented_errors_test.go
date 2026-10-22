package errors

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Problem responses use the RFC 7807 media type and fields.
func TestProblemDetailsUsesRFC7807Response(t *testing.T) {
	p := NewProblem(ErrorTypeConflict, "conflict", http.StatusConflict, "Conflict", "Already exists")
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	if err := p.Write(w, r); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusConflict || w.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("response=%d %v", w.Code, w.Header())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"type", "title", "status", "detail"} {
		if _, ok := body[key]; !ok {
			t.Fatalf("missing %s: %v", key, body)
		}
	}
}

// A key can be validated and consistently hashed for replay lookup.
func TestIdempotencyKeyProducesStableStorageKey(t *testing.T) {
	first := ExtractIdempotencyKey("Idempotency-Key", "purchase-42")
	second := ExtractIdempotencyKey("Idempotency-Key", "purchase-42")
	if !first.Valid || !second.Valid || first.Key != second.Key || first.Key == "purchase-42" {
		t.Fatalf("keys=%+v %+v", first, second)
	}
	if ExtractIdempotencyKey("Idempotency-Key", "").Valid {
		t.Fatal("empty key accepted")
	}
}
