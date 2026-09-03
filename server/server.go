package server

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net"
	"os"
	"time"

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
	slog.Info("server is ready for requests")
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
	defer conn.Close()
	slog.Info("connection received, now sleeping")
	if tlsConn, ok := conn.(*tls.Conn); ok {
		hctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tlsConn.HandshakeContext(hctx); err != nil {
			s.logger.Warn("tls handshake failed", "remote", conn.RemoteAddr(), "err", err)
			return
		}
		s.logger.Info("tls handshake ok", "remote", conn.RemoteAddr())
		// TODO: lb.Next() for target
		// TODO: resilience.Allos(target)
		// TODO: transport.pump(conn, upstream)
		// TODO: resilience.Record + telemetry on close

	}

	time.Sleep(time.Second * 5)
	slog.Info("Connection completed Terminating ")
}
