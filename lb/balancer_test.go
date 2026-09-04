package lb

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func servers(addr string, ready chan<- struct{}) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("unable to create server for port", "port", addr, "error", err)
	}
	close(ready)
	for {
		slog.Info("server started listening on ", "address", addr)
		fmt.Println("server started listening")
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("error while accepting connection on ", "address ", addr, "error", err)
			return
		}
		slog.Info("connection received on ", "address ", addr)
		conn.Close()
	}
}

func call(roundRobin *RoundRobin) {
	upstream, err := (*roundRobin).Next()
	if err != nil {
		slog.Error("unable to get the upstream")
		return
	}
	address := (*upstream).Addr
	conn, err := net.Dial("tcp", address)
	defer conn.Close()
	if err != nil {
		slog.Error("unable to dial server with ", "address ", (*upstream).Addr)
	}
}

func TestConcConn(t *testing.T) {
	t.Run("running 100 concurrent request to 5 servers", func(t *testing.T) {
		addr := 8080
		var upstreamServer []*Upstream
		for i := range 5 {
			ready := make(chan struct{})
			address := fmt.Sprintf(":%d", addr+i)
			go servers(address, ready)
			<-ready
			upstreamServer = append(upstreamServer, &Upstream{Addr: address})
		}
		var wg sync.WaitGroup
		roundRobin := RoundRobin{upstream: upstreamServer}
		for range 5 {
			wg.Add(1)
			go call(&roundRobin)
		}
		wg.Wait()
	})
}

func TestLeastConnection(t *testing.T) {
	t.Run("testing the Least Connection working", func(t *testing.T) {
		var upstream []*Upstream
		address := 8080
		var counter atomic.Uint64
		for i := range 3 {
			counter.Store(uint64(i))
			addr := fmt.Sprintf(":%d", address+i)
			upstreamServer := Upstream{Addr: addr, ActiveConn: counter}
			upstream = append(upstream, &upstreamServer)
		}
		leastCOnn := LeastConnections{upstream}
		finalUpstream, err := leastCOnn.Next()
		if err != nil {
			slog.Error("unable to get the upstream", "error", err)
			return
		}
		assert.Equal(t, ":8080", (*finalUpstream).Addr)
	})
}

func LeastConnCall(leastConnection *LeastConnections, wg *sync.WaitGroup) {
	defer wg.Done()
	upstream, err := leastConnection.Next()
	if err != nil {
		slog.Error("unable to call next on least Connection struct ", "error", err)
		return
	}
	conn, err := net.Dial("tcp", (*upstream).Addr)
	if err != nil {
		slog.Error("unable to dial into server ", "addressl", (*upstream).Addr)
		return
	}
	defer conn.Close()
	defer leastConnection.Release(upstream)
}

func TestLeastConnConc(t *testing.T) {
	t.Run("testing Least connection concurrently", func(t *testing.T) {
		addr := 8080
		var upstreamServer []*Upstream
		var counter atomic.Uint64
		for i := range 5 {
			counter.Store(uint64(5 - i))
			ready := make(chan struct{})
			address := fmt.Sprintf(":%d", addr+i)
			go servers(address, ready)
			<-ready
			upstreamServer = append(upstreamServer, &Upstream{Addr: address, ActiveConn: counter})
		}
		var wg sync.WaitGroup
		leastConn := LeastConnections{upstreamServer}
		for range 5 {
			wg.Add(1)
			go LeastConnCall(&leastConn, &wg)
		}
		wg.Wait()
	})
}
