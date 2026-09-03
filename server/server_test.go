package server

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"net"
	"os"
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

func TestTlsHandshake(t *testing.T) {
	t.Run("testing tls handshake with server", func(t *testing.T) {
		addr := ":8080"
		maxConns := 2
		var wg sync.WaitGroup
		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
		servercert, err := tls.LoadX509KeyPair("/home/ghozt/go/Projects/cert/server.crt", "/home/ghozt/go/Projects/cert/server.key")
		if err != nil {
			slog.Error("unabel to load server certs", "err", err)
		}
		caCert, err := os.ReadFile("/home/ghozt/go/Projects/cert/ca.crt")
		if err != nil {
			slog.Error("unable to load ca.cert file", "err", err)
		}
		servercaPool := x509.NewCertPool()
		servercaPool.AppendCertsFromPEM(caCert)
		servertlsConfig := &tls.Config{
			Certificates: []tls.Certificate{servercert},
			ClientCAs:    servercaPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
			MinVersion:   tls.VersionTLS13,
		}
		server, err := New(addr, maxConns, ctx, servertlsConfig)
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
		cert, err := tls.LoadX509KeyPair("/home/ghozt/go/Projects/cert/client.crt", "/home/ghozt/go/Projects/cert/client.key")
		if err != nil {
			slog.Error("unabel to load client certs", "err", err)
		}
		caPool := x509.NewCertPool()
		caPool.AppendCertsFromPEM(caCert)
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caPool,
			MinVersion:   tls.VersionTLS13,
		}
		tlsConfig.ServerName = "localhost"
		_, err = tls.Dial("tcp", "localhost:8080", tlsConfig)
		wg.Wait()
		assert.Equal(t, nil, err)
	})
}

func TestExpiredCertF(t *testing.T) {
	t.Run("testing client which provices expired certificate", func(t *testing.T) {
		addr := ":8080"
		maxConns := 2
		var wg sync.WaitGroup
		ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
		servercert, err := tls.LoadX509KeyPair("/home/ghozt/go/Projects/cert/server.crt", "/home/ghozt/go/Projects/cert/server.key")
		if err != nil {
			slog.Error("unabel to load server certs", "err", err)
		}
		caCert, err := os.ReadFile("/home/ghozt/go/Projects/cert/ca.crt")
		if err != nil {
			slog.Error("unable to load ca.cert file", "err", err)
		}
		servercaPool := x509.NewCertPool()
		servercaPool.AppendCertsFromPEM(caCert)
		servertlsConfig := &tls.Config{
			Certificates: []tls.Certificate{servercert},
			ClientCAs:    servercaPool,
			ClientAuth:   tls.RequireAndVerifyClientCert,
			MinVersion:   tls.VersionTLS13,
		}
		server, err := New(addr, maxConns, ctx, servertlsConfig)
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
		cert, err := tls.LoadX509KeyPair("/home/ghozt/go/Projects/cert/expired.crt", "/home/ghozt/go/Projects/cert/client.key")
		if err != nil {
			slog.Error("unabel to load client certs", "err", err)
		}
		caPool := x509.NewCertPool()
		caPool.AppendCertsFromPEM(caCert)
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caPool,
			MinVersion:   tls.VersionTLS13,
		}
		tlsConfig.ServerName = "localhost"
		_, err = tls.Dial("tcp", "localhost:8080", tlsConfig)
		slog.Info("error of the tls.Dial", "err", err)
		wg.Wait()
	})
}
