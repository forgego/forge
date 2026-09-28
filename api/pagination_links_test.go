package api

import (
	"crypto/tls"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildPaginatedResponse_Links(t *testing.T) {
	t.Run("preserves query params and builds valid next and previous", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/api/products?page=2&page_size=10&search=shoe&ordering=-price", nil)
		resp := BuildPaginatedResponse(req, []string{}, 35, 2, 10)

		require.NotNil(t, resp.Next)
		assert.Equal(t, "http://example.com/api/products?ordering=-price&page=3&page_size=10&search=shoe", *resp.Next)

		require.NotNil(t, resp.Previous)
		assert.Equal(t, "http://example.com/api/products?ordering=-price&page=1&page_size=10&search=shoe", *resp.Previous)
	})

	t.Run("TLS request produces https link", func(t *testing.T) {
		req := httptest.NewRequest("GET", "http://example.com/api/products?page=2&page_size=10", nil)
		req.TLS = &tls.ConnectionState{}
		resp := BuildPaginatedResponse(req, []string{}, 35, 2, 10)

		require.NotNil(t, resp.Next)
		assert.Equal(t, "https://example.com/api/products?page=3&page_size=10", *resp.Next)

		require.NotNil(t, resp.Previous)
		assert.Equal(t, "https://example.com/api/products?page=1&page_size=10", *resp.Previous)
	})

	t.Run("first page previous is nil and last page next is nil", func(t *testing.T) {
		tests := []struct {
			name     string
			page     int
			pageSize int
			total    int
			wantNext bool
			wantPrev bool
		}{
			{
				name:     "first page",
				page:     1,
				pageSize: 10,
				total:    35,
				wantNext: true,
				wantPrev: false,
			},
			{
				name:     "last page",
				page:     4,
				pageSize: 10,
				total:    35,
				wantNext: false,
				wantPrev: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest("GET", "http://example.com/api/products", nil)
				resp := BuildPaginatedResponse(req, []string{}, tt.total, tt.page, tt.pageSize)

				if tt.wantNext {
					assert.NotNil(t, resp.Next)
				} else {
					assert.Nil(t, resp.Next)
				}

				if tt.wantPrev {
					assert.NotNil(t, resp.Previous)
				} else {
					assert.Nil(t, resp.Previous)
				}
			})
		}
	})
}
