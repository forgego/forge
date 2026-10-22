package api

import (
	"net/http/httptest"
	"testing"

	"github.com/forgego/forge/api/renderers"
)

// Page-number parameters select the correct result offset.
func TestPaginationUsesPageAndPageSize(t *testing.T) {
	r := httptest.NewRequest("GET", "/products?page=3&page_size=5", nil)
	page, size, offset := ParsePaginationParams(r, 20)
	if page != 3 || size != 5 || offset != 10 {
		t.Fatalf("page=%d size=%d offset=%d", page, size, offset)
	}
	response := BuildPaginatedResponse(r, []string{"eleven"}, 21, page, size)
	if response.Next == nil || response.Previous == nil {
		t.Fatalf("missing page links: %+v", response)
	}
}

// The Accept header selects an installed representation.
func TestContentNegotiationSelectsAcceptedRenderer(t *testing.T) {
	cn := NewContentNegotiator([]renderers.Renderer{renderers.NewJSONRenderer(), renderers.NewXMLRenderer()}, nil)
	r := httptest.NewRequest("GET", "/products", nil)
	r.Header.Set("Accept", "application/xml")
	if got := cn.SelectRenderer(r).MediaType(); got != "application/xml" {
		t.Fatalf("renderer=%q", got)
	}
}
