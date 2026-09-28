package netutil

import (
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientIP(t *testing.T) {
	defaultTrustedPrefix := netip.MustParsePrefix("10.0.0.0/8")

	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		xRealIP    string
		trusted    []netip.Prefix
		expectedIP string
	}{
		{
			name:       "untrusted peer, header ignored",
			remoteAddr: "203.0.113.9:5555",
			xff:        "1.2.3.4",
			trusted:    []netip.Prefix{defaultTrustedPrefix},
			expectedIP: "203.0.113.9",
		},
		{
			name:       "untrusted peer, X-Real-IP ignored",
			remoteAddr: "203.0.113.9:5555",
			xRealIP:    "1.2.3.4",
			trusted:    []netip.Prefix{defaultTrustedPrefix},
			expectedIP: "203.0.113.9",
		},
		{
			name:       "rightmost untrusted",
			remoteAddr: "10.0.0.5:80",
			xff:        "1.2.3.4, 198.51.100.7",
			trusted:    []netip.Prefix{defaultTrustedPrefix},
			expectedIP: "198.51.100.7",
		},
		{
			name:       "rightmost trusted skipped",
			remoteAddr: "10.0.0.5:80",
			xff:        "1.2.3.4, 10.0.0.9",
			trusted:    []netip.Prefix{defaultTrustedPrefix},
			expectedIP: "1.2.3.4",
		},
		{
			name:       "garbage entry skipped",
			remoteAddr: "10.0.0.5:80",
			xff:        "garbage, 198.51.100.7",
			trusted:    []netip.Prefix{defaultTrustedPrefix},
			expectedIP: "198.51.100.7",
		},
		{
			name:       "no XFF, X-Real-IP present",
			remoteAddr: "10.0.0.5:80",
			xRealIP:    "198.51.100.8",
			trusted:    []netip.Prefix{defaultTrustedPrefix},
			expectedIP: "198.51.100.8",
		},
		{
			name:       "XFF only has trusted IP",
			remoteAddr: "10.0.0.5:80",
			xff:        "10.1.1.1",
			trusted:    []netip.Prefix{defaultTrustedPrefix},
			expectedIP: "10.0.0.5",
		},
		{
			name:       "IPv6 trusted peer",
			remoteAddr: "[::1]:80",
			xff:        "2001:db8::1",
			trusted:    []netip.Prefix{netip.MustParsePrefix("::1/128")},
			expectedIP: "2001:db8::1",
		},
		{
			name:       "no port in RemoteAddr",
			remoteAddr: "203.0.113.9",
			trusted:    []netip.Prefix{defaultTrustedPrefix},
			expectedIP: "203.0.113.9",
		},
		{
			name:       "trusted empty",
			remoteAddr: "10.0.0.5:80",
			xff:        "1.2.3.4",
			trusted:    nil,
			expectedIP: "10.0.0.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			ip := ClientIP(req, tt.trusted)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

func TestSetTrustedProxies(t *testing.T) {
	// Reset to empty
	err := SetTrustedProxies(nil)
	require.NoError(t, err)
	assert.Empty(t, TrustedProxies())

	// Set valid proxies
	err = SetTrustedProxies([]string{"10.0.0.0/8", "127.0.0.1", "::1"})
	require.NoError(t, err)
	initial := TrustedProxies()
	require.Len(t, initial, 3)

	// Set invalid proxies: must be all-or-nothing
	err = SetTrustedProxies([]string{"10.0.0.0/8", "bad"})
	assert.Error(t, err)
	assert.Equal(t, initial, TrustedProxies(), "TrustedProxies must remain unchanged on invalid entry")
}
