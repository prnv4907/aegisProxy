package server

import (
	"context"
	"log/slog"
	"net"
	"os"

	"golang.org/x/sync/semaphore"
)

type server struct {
	listner net.Listener
	sem     *semaphore.Weighted
	logger  *slog.Logger
}

func New(add string, maxConns int) (*server, error) {
	listner, err := net.Listen("tcp", add)
	if err != nil {
		return nil, err
	}
	semChan := semaphore.NewWeighted(int64(maxConns))
	slog := slog.New(slog.NewTextHandler(os.Stderr, nil))
	return &server{listner, semChan, slog}, nil
}

func (s *server) Server() error {
	for true {
		conn, err := s.listner.Accept()
		if err != nil {
			return err
		}
		ctx := context.Context(context.Background())
		_ = s.sem.Acquire(ctx, 1)
		go s.handleConn(conn)
	}
	return nil
}

func (s *server) handleConn(conn net.Conn) {
}
