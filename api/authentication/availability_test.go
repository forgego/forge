package authentication

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type sessionLookupFunc func(*http.Request) (interface{}, error)

func (f sessionLookupFunc) GetUser(r *http.Request) (interface{}, error) { return f(r) }

func TestSessionAuthenticationUsesSessionUser(t *testing.T) {
	user := &struct{ Name string }{"alice"}
	a := NewSessionAuthentication(sessionLookupFunc(func(r *http.Request) (interface{}, error) {
		if _, err := r.Cookie("session"); err != nil {
			t.Fatal(err)
		}
		return user, nil
	}))
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.AddCookie(&http.Cookie{Name: "session", Value: "valid"})
	result, err := a.Authenticate(r)
	if err != nil || result == nil || result.User != user {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestBasicAuthenticationUsesDecodedCredentials(t *testing.T) {
	a := NewBasicAuthentication(func(name, password string) (interface{}, error) {
		if name != "alice" || password != "secret" {
			t.Fatalf("credentials=%q:%q", name, password)
		}
		return "alice", nil
	})
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.SetBasicAuth("alice", "secret")
	result, err := a.Authenticate(r)
	if err != nil || result == nil || result.User != "alice" {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestAPIKeyAuthenticationUsesHeaderAndQueryKey(t *testing.T) {
	a := NewAPIKeyAuthentication(func(key string) (interface{}, error) { return key, nil })
	for _, tc := range []struct{ header, query, want string }{
		{"header-key", "query-key", "header-key"}, {"", "query-key", "query-key"},
	} {
		r := httptest.NewRequest(http.MethodGet, "/?api_key="+tc.query, nil)
		r.Header.Set("X-API-Key", tc.header)
		result, err := a.Authenticate(r)
		if err != nil || result == nil || result.User != tc.want {
			t.Fatalf("result=%+v err=%v", result, err)
		}
	}
}
