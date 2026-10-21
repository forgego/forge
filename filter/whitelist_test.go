package filter

import (
	"context"
	"github.com/forgego/forge/orm"
	"net/http/httptest"
	"testing"
)

func TestConfiguredFilterSecurityIsRequestWhitelist(t *testing.T) {
	fs, err := NewFilterSet[MockModel]()
	if err != nil {
		t.Fatal(err)
	}
	qs, err := orm.NewQuerySet[MockModel]("mock_models")
	if err != nil {
		t.Fatal(err)
	}
	security := DefaultSecurityConfig()
	security.AllowedFields["mock_models"] = []string{"username"}
	security.AllowedLookups["username"] = []string{"exact"}
	fs.WithQueryset(qs).WithSecurity(security)
	if _, err := fs.ApplyQueryParams(context.Background(), httptest.NewRequest("GET", "/?username=ada", nil)); err != nil {
		t.Fatal(err)
	}
	if _, err := fs.ApplyQueryParams(context.Background(), httptest.NewRequest("GET", "/?email=hidden", nil)); err == nil {
		t.Fatal("unlisted field admitted")
	}
}
