package service

import (
	identityconfig "github.com/forgego/forge/identity/config"
	"github.com/forgego/forge/identity/models"
	"testing"
)

// Default policy rejects weak passwords and accepts a strong one.
func TestPasswordPolicyRejectsShortAndMissingClasses(t *testing.T) {
	svc := NewPasswordService(nil, nil, identityconfig.DefaultPasswordPolicy())
	u := &models.User{Username: "owner", Email: "owner@example.invalid"}
	for _, password := range []string{"short", "alllowercase123", "ALLUPPERCASE123", "NoDigitsHere"} {
		if err := svc.ValidatePassword(password, u); err == nil {
			t.Fatalf("accepted weak password %q", password)
		}
	}
	if err := svc.ValidatePassword("LongStrongPass123", u); err != nil {
		t.Fatal(err)
	}
}
