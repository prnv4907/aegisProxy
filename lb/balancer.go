package lb

import (
	"errors"
	"fmt"
	"hash/crc32"
	"slices"
	"sync"
	"sync/atomic"
)

var (
	ErrNoHealthyUpstream = errors.New("no healty upstream available")
	ErrEmptyChain        = errors.New("no upstream present in chain")
)

type Balancer interface {
	Next(s string) (*Upstream, error) // this needs to be goroutine safe as many goroutine (for each connection ) will try to use it to send traffic to the server
}
type Upstream struct {
	Addr       string
	_          [48]byte
	ActiveConn atomic.Uint64
}

type RoundRobin struct {
	upstream []*Upstream
	_        [56]byte
	counter  atomic.Uint64
}

func (r *RoundRobin) Next(s string) (*Upstream, error) {
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

func (l *LeastConnections) Next(s string) (*Upstream, error) {
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

type ConsistentHash struct {
	ring       map[uint64]*Upstream
	sortedKeys []uint64
	mu         sync.RWMutex
}

func New() *ConsistentHash {
	c := &ConsistentHash{
		ring: make(map[uint64]*Upstream),
	}
	return c
}

func (c *ConsistentHash) IsEmpty() bool {
	return len(c.ring) == 0
}

func (c *ConsistentHash) Add(u *Upstream) {
	addr := u.Addr
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range 3 {
		serverKey := fmt.Sprintf("%s server %d", addr, i)
		hash := uint64(crc32.ChecksumIEEE([]byte(serverKey)))
		c.ring[hash] = u
		c.sortedKeys = append(c.sortedKeys, hash)
		slices.Sort(c.sortedKeys)
	}
}

func (c *ConsistentHash) Remove(u *Upstream) error {
	if c.IsEmpty() {
		return ErrEmptyChain
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, up := range c.ring {
		if up == u {
			delete(c.ring, key)
			for i, kk := range c.sortedKeys {
				if kk == key {
					c.sortedKeys = append(c.sortedKeys[:i], c.sortedKeys[i+1:]...)
					break
				}
			}
		}
	}
	return nil
}

func (c *ConsistentHash) Next(host string) (*Upstream, error) {
	if c.IsEmpty() {
		return nil, ErrEmptyChain
	}
	hash := uint64(crc32.ChecksumIEEE([]byte(host)))
	c.mu.RLock()
	defer c.mu.RUnlock()
	firstKey := c.sortedKeys[0]
	for _, key := range c.sortedKeys {
		if key > hash {
			firstKey = key
			return c.ring[firstKey], nil
		}
	}
	return c.ring[firstKey], nil
}
