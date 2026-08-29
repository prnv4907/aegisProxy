package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"

	"golang.org/x/sync/semaphore"
)

type Server struct {
	listner net.Listener
	sem     *semaphore.Weighted
	logger  *slog.Logger
	ctx     context.Context
}

func New(addr string, maxConns int, ctx context.Context, tlsConfig *tls.Config) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	if tlsConfig != nil {
		listener = tls.NewListener(listener, tlsConfig)
	}
	semChan := semaphore.NewWeighted(int64(maxConns))
	Log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	return &Server{listener, semChan, Log, ctx}, nil
}

func (s *Server) serve() error {
	go func() {
		<-s.ctx.Done()
		s.listner.Close()
	}()
	for {
		conn, err := s.listner.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		ctx := context.Context(context.Background())
		_ = s.sem.Acquire(ctx, 1)
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer s.sem.Release(1)
	for {
		fmt.Println("bello ")
	}
}
