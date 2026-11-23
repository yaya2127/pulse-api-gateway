package proxy

import (
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
)

var ErrNoAvailableUpstream = errors.New("no alive upstream available in pool")

type Balancer interface {
	Next(r *http.Request) (*Upstream, error)
	GetUpstreams() []*Upstream
	AddUpstream(u *Upstream)
}
