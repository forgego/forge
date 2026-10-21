package errors

import (
	"strings"
	"testing"
)

func TestRegisteredErrorCodeCanBeRetrieved(t *testing.T) {
	code := &ErrorCode{Code: "TEST_REGISTRY_CONTRACT", Description: "sample", HTTPStatus: 422, Type: ErrorTypeValidation}
	if err := RegisterErrorCode(code); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { globalRegistry.mu.Lock(); delete(globalRegistry.codes, code.Code); globalRegistry.mu.Unlock() })
	got, ok := GetErrorCode(code.Code)
	if !ok || got != code {
		t.Fatalf("code=%+v found=%v", got, ok)
	}
	if err := RegisterErrorCode(code); err == nil {
		t.Fatal("duplicate registration succeeded")
	}
}

func TestSetTypeBaseURLChangesProblemType(t *testing.T) {
	previous := GetTypeBaseURL()
	t.Cleanup(func() { SetTypeBaseURL(previous) })
	SetTypeBaseURL("https://example.test/problems")
	got := NewBaseError(ErrorTypeValidation, "INVALID", 400, "Invalid", "bad").ToProblem().Type
	if !strings.HasPrefix(got, "https://example.test/problems/") {
		t.Fatalf("problem type=%q", got)
	}
}
