package transport

import (
	"net"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestPipe(t *testing.T) {
	t.Run("testing Pipe function", func(t *testing.T) {
		Client, proxyClient := net.Pipe()
		Backend, proxyBackend := net.Pipe()
		go Pipe(proxyClient, proxyBackend)
		go func() {
			Client.Write([]byte("hey there"))
		}()
		buf := make([]byte, 1024)
		n, err := Backend.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		got := string(buf[:n])
		assert.Equal(t, "hey there", got)
		Client.Close()
		Backend.Close()
	})
}
