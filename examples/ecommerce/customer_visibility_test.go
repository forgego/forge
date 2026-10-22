package main

import (
	"net/http"
	"strconv"
	"testing"
)

func TestCustomerAPIHidesCredentialFields(t *testing.T) {
	router, _ := setupTestEcommerceRouter(t)
	created := performJSONRequest(t, router, http.MethodPost, "/api/v1/customers/", map[string]interface{}{
		"email":         "customer@example.invalid",
		"password_hash": "test-only-placeholder",
		"first_name":    "Test",
		"last_name":     "Customer",
	}, http.StatusCreated)
	for _, field := range []string{"password_hash", "verification_token", "reset_password_token"} {
		if _, visible := created[field]; visible {
			t.Errorf("create response includes %s", field)
		}
	}
	id := int64(created["id"].(float64))
	got := performJSONRequest(t, router, http.MethodGet, "/api/v1/customers/"+strconv.FormatInt(id, 10), nil, http.StatusOK)
	for _, field := range []string{"password_hash", "verification_token", "reset_password_token"} {
		if _, visible := got[field]; visible {
			t.Errorf("retrieve response includes %s", field)
		}
	}
}
