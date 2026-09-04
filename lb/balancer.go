package lb

import (
	"errors"
	"sync/atomic"
)

var ErrNoHealthyUpstream = errors.New("no healty upstream available")

type Upstream struct {
	Addr       string
	_          [48]byte
	ActiveConn atomic.Uint64
}
type Balancer interface {
	Next() (*Upstream, error) // this needs to be goroutine safe as many goroutine (for each connection ) will try to use it to send traffic to the server
}

type RoundRobin struct {
	upstream []*Upstream
	_        [56]byte
	counter  atomic.Uint64
}

func (r *RoundRobin) Next() (*Upstream, error) {
	if len(r.upstream) == 0 {
		return nil, ErrNoHealthyUpstream
	}
	roundRobinInt := r.counter.Add(1) - 1
	NextUpstream := r.upstream[int(roundRobinInt)%len(r.upstream)]
	return NextUpstream, nil
}

type LeastConnections struct {
	upstream []*Upstream
}

func (l *LeastConnections) Next() (*Upstream, error) {
	upLen := len(l.upstream)
	if upLen == 0 {
		return nil, ErrNoHealthyUpstream
	}
	index := 0
	value := l.upstream[0].ActiveConn.Load()
	if upLen > 1 {
		for i := 1; i < upLen; i++ {
			currConn := l.upstream[i].ActiveConn.Load()
			if value > currConn {
				value = currConn
				index = i
			}
		}
	}
	l.upstream[index].ActiveConn.Add(1)
	return l.upstream[index], nil
}

func (l LeastConnections) Release(u *Upstream) {
	u.ActiveConn.Add(^uint64(0))
}
