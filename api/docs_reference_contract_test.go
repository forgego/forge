package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/forgego/forge/api/authentication"
	"github.com/forgego/forge/api/permissions"
)

type referenceSerializer struct{ *BaseSerializer }

func (s *referenceSerializer) New() Serializer {
	return &referenceSerializer{NewBaseSerializer(nil)}
}

func (s *referenceSerializer) Validate() error {
	s.AddError("price", "price must be positive")
	return errors.New("invalid price")
}

func TestSerializerReferenceContract(t *testing.T) {
	t.Run("model serializer implements contract", func(t *testing.T) {
		var serializer Serializer = NewModelSerializer(struct{ Name string }{Name: "item"})
		if serializer == nil {
			t.Fatal("model serializer is nil")
		}
	})
	t.Run("custom serializer validation and cloning", func(t *testing.T) {
		var validate func(*referenceSerializer) error = (*referenceSerializer).Validate
		s := &referenceSerializer{NewBaseSerializer(nil)}
		if err := validate(s); err == nil || len(s.Errors()["price"]) != 1 {
			t.Fatalf("validation did not report field error: %v, %+v", err, s.Errors())
		}
		if _, ok := s.New().(*referenceSerializer); !ok {
			t.Fatal("new serializer lost custom validation")
		}
	})
}

func TestViewSetReferenceErrorResponses(t *testing.T) {
	vs := NewBaseViewSet(func() Serializer { return &referenceSerializer{NewBaseSerializer(nil)} }, &requiredScalarManager{}, &requiredScalarRequestModel{})
	vs.Authentication = []authentication.Authentication{}
	vs.Permissions = []permissions.Permission{}
	for _, tc := range []struct {
		name, body string
		wantDetail bool
	}{
		{"malformed JSON", `{`, true},
		{"serializer validation", `{"count":1,"enabled":true}`, false},
	} {
		t.Run("create rejects "+tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			vs.Create(rec, httptest.NewRequest(http.MethodPost, "/items/", strings.NewReader(tc.body)))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
			}
			var body map[string]any
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if tc.wantDetail {
				if body["detail"] != "Invalid JSON" {
					t.Fatalf("problem detail = %v", body["detail"])
				}
			} else if body["errors"] == nil {
				t.Fatalf("field errors missing: %v", body)
			}
		})
	}
}
