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
