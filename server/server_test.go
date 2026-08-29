package server

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gookit/goutil/testutil/assert"
)

func TestServerError(t *testing.T) {
	t.Run("testing whether server errors out of panics", func(t *testing.T) {
		addr := ":8080"
		maxConns := 3
		ctx, _ := context.WithTimeout(context.Background(), time.Second)
		server, err := New(addr, maxConns, ctx)
		if err != nil {
			fmt.Printf("unable to make a server %v\n", err)
			return
		}
		err = server.serve()
		assert.Equal(t, err, nil)
	})
}
