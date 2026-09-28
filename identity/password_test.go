package identity

import (
	"testing"
)

func TestPasswordHelpers(t *testing.T) {
	plain := "secret-password-123"

	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if !CheckPassword(plain, hash) {
		t.Fatal("CheckPassword returned false for valid password")
	}

	if !CheckPasswordHash(plain, hash) {
		t.Fatal("CheckPasswordHash returned false for valid password")
	}

	if !IsHashed(hash) {
		t.Fatal("IsHashed returned false for valid hash")
	}

	if IsHashed(plain) {
		t.Fatal("IsHashed returned true for plaintext")
	}

	if NeedsRehash(hash, DefaultCost) {
		t.Fatal("NeedsRehash returned true for hash with DefaultCost")
	}
}
