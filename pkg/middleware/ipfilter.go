package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
)

type IPFilterMode int

const (
	ModeAllowlist IPFilterMode = iota
	ModeBlocklist
)

type IPFilter struct {
	mode        IPFilterMode
	subnets     []*net.IPNet
	trustedCIDR []*net.IPNet
	mu          sync.RWMutex
}

func NewIPFilter(mode IPFilterMode, cidrs []string) (*IPFilter, error) {
	subnets := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		subnets = append(subnets, ipNet)
	}
	return &IPFilter{mode: mode, subnets: subnets}, nil
}

func (f *IPFilter) IsAllowed(ip net.IP) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	matched := false
	for _, netw := range f.subnets {
		if netw.Contains(ip) {
			matched = true
			break
		}
	}

	if f.mode == ModeAllowlist {
		return matched
	}
	return !matched
}

func (f *IPFilter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}

		ip := net.ParseIP(strings.TrimSpace(host))
		if ip == nil || !f.IsAllowed(ip) {
			http.Error(w, "Forbidden: IP rejected by security gateway policy", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
