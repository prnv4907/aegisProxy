package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/gookit/goutil/testutil/assert"
)

func TestServerError(t *testing.T) {
	t.Run("testing whether server errors out of panics", func(t *testing.T) {
		addr := ":8080"
		maxConns := 3
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		server, err := New(addr, maxConns, ctx, nil)
		if err != nil {
			fmt.Printf("unable to make a server %v\n", err)
			return
		}
		err = server.serve()
		assert.Equal(t, err, nil)
	})
}

func TestOverLoadingConn(t *testing.T) {
	t.Run("testing 5 concurrent connections with maxAllowed set to 2 ", func(t *testing.T) {
		addr := ":8080"
		maxConns := 2
		var wg sync.WaitGroup
		ctx, _ := context.WithTimeout(context.Background(), time.Second*20)
		server, err := New(addr, maxConns, ctx, nil)
		if err != nil {
			slog.Info("unable to make a server", "err", err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := server.serve()
			if err != nil {
				slog.Info("server errored out", "err", err)
			}
		}()
		if err != nil {
			slog.Info("unable to server the server", "err", err)
		}
		for range 5 {
			net.Dial("tcp", addr)
		}
		wg.Wait()
	})
}

func TestFlooding(t *testing.T) {
	t.Run("flooding the server with 100 concurrent connections", func(t *testing.T) {
		addr := ":8080"
		maxConns := 2
		var wg sync.WaitGroup
		ctx, _ := context.WithTimeout(context.Background(), time.Second*20)
		server, err := New(addr, maxConns, ctx, nil)
		if err != nil {
			slog.Info("unable to make a server", "err", err)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := server.serve()
			if err != nil {
				slog.Info("server errored out", "err", err)
			}
		}()
		if err != nil {
			slog.Info("unable to server the server", "err", err)
		}
		for range 100 {
			net.Dial("tcp", addr)
		}
		wg.Wait()
	})
}
