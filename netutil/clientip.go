package netutil

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync/atomic"
)

var trustedProxies atomic.Pointer[[]netip.Prefix]

func init() {
	empty := make([]netip.Prefix, 0)
	trustedProxies.Store(&empty)
}

// SetTrustedProxies configures which peers may supply X-Forwarded-For / X-Real-IP.
// Accepts CIDRs ("10.0.0.0/8") or bare IPs ("127.0.0.1"). Empty = trust no one (default).
func SetTrustedProxies(entries []string) error {
	prefixes := make([]netip.Prefix, 0, len(entries))
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if strings.Contains(entry, "/") {
			prefix, err := netip.ParsePrefix(entry)
			if err != nil {
				return err
			}
			prefixes = append(prefixes, prefix.Masked())
		} else {
			addr, err := netip.ParseAddr(entry)
			if err != nil {
				return err
			}
			addr = addr.Unmap()
			prefixes = append(prefixes, netip.PrefixFrom(addr, addr.BitLen()))
		}
	}
	trustedProxies.Store(&prefixes)
	return nil
}

// TrustedProxies returns the current list (safe for concurrent use).
func TrustedProxies() []netip.Prefix {
	p := trustedProxies.Load()
	if p == nil {
		return nil
	}
	return *p
}

// ClientIP returns the best-effort client IP for rate limiting.
func ClientIP(r *http.Request, trusted []netip.Prefix) string {
	peerStr, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		peerStr = r.RemoteAddr
	}

	cleanPeer := strings.TrimPrefix(strings.TrimSuffix(peerStr, "]"), "[")
	peerAddr, err := netip.ParseAddr(cleanPeer)
	if err == nil {
		peerAddr = peerAddr.Unmap()
	}

	peerResult := peerStr
	if peerAddr.IsValid() {
		peerResult = peerAddr.String()
	}

	if !peerAddr.IsValid() || !isTrusted(peerAddr, trusted) {
		return peerResult
	}

	// Peer is trusted: split X-Forwarded-For on ",", trim spaces, walk from the RIGHT;
	// return the first entry that parses as an IP and is NOT in a trusted prefix.
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		var parts []string
		if vals := r.Header.Values("X-Forwarded-For"); len(vals) > 0 {
			for _, v := range vals {
				for _, part := range strings.Split(v, ",") {
					part = strings.TrimSpace(part)
					if part != "" {
						parts = append(parts, part)
					}
				}
			}
		} else {
			for _, part := range strings.Split(xff, ",") {
				part = strings.TrimSpace(part)
				if part != "" {
					parts = append(parts, part)
				}
			}
		}

		for i := len(parts) - 1; i >= 0; i-- {
			part := parts[i]
			cleanPart := strings.TrimPrefix(strings.TrimSuffix(part, "]"), "[")
			addr, err := netip.ParseAddr(cleanPart)
			if err != nil {
				continue
			}
			addr = addr.Unmap()
			if !isTrusted(addr, trusted) {
				return addr.String()
			}
		}
	}

	// If none found: if X-Real-IP parses as an IP, return it.
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		xri = strings.TrimSpace(xri)
		cleanXRI := strings.TrimPrefix(strings.TrimSuffix(xri, "]"), "[")
		if addr, err := netip.ParseAddr(cleanXRI); err == nil {
			return addr.Unmap().String()
		}
	}

	// Otherwise return peer.
	return peerResult
}

func isTrusted(addr netip.Addr, trusted []netip.Prefix) bool {
	if !addr.IsValid() {
		return false
	}
	for _, prefix := range trusted {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}
