package transport

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

var bufPool sync.Pool = sync.Pool{
	New: func() any {
		buf := make([]byte, 32*1024)
		return &buf
	},
}

func Pipe(client, backend net.Conn) error {
	errChan := make(chan error, 2)
	go copyDirection(backend, client, "client->backend", errChan)
	go copyDirection(client, backend, "backend->client", errChan)
	err := <-errChan
	log.Printf("connection closed first by: %v", err)
	client.Close()
	backend.Close()
	<-errChan
	return err
}

func copyDirection(dist, src net.Conn, label string, errChan chan<- error) {
	buff := bufPool.Get().(*[]byte)
	defer bufPool.Put(&buff)
	_, err := io.CopyBuffer(dist, src, *buff)
	errChan <- fmt.Errorf("%s : %w", label, err)
}
