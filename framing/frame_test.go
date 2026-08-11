package framing

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/gookit/goutil/testutil/assert"
)

func TestEncoding(t *testing.T) {
	t.Run("testing encoding ", func(t *testing.T) {
		task := []byte{1}
		payload := "heythere"
		payloadBytes := ([]byte)(payload)
		frame := Frame{task[0], payloadBytes}
		buff, err := frame.Encode()
		if err != nil {
			t.Errorf("error in endoding %s", err)
		}
		fmt.Println(buff)
	})
}

func TestDecoding(t *testing.T) {
	t.Run("testing Decoding", func(t *testing.T) {
		task := []byte{1}
		payload := []byte("hello there")
		len := len(payload)
		buf := make([]byte, 1+4+len)
		buf[0] = task[0]
		binary.BigEndian.PutUint32(buf[1:5], uint32(len))
		for i := range len {
			buf[5+i] = payload[i]
		}
		reader := bytes.NewReader(buf)
		frame := Frame{}
		finalFrame, err := frame.Decoder(reader)
		if err != nil {
			t.Errorf("error in decoding the frame")
		}
		assert.Equal(t, payload, finalFrame.Payload)
	})
}

func TestFullFlow(t *testing.T) {
	t.Run("testing full encoding decoding flow", func(t *testing.T) {
		task := []byte{1}
		payload := "heythere"
		payloadBytes := ([]byte)(payload)
		frame := Frame{task[0], payloadBytes}
		buff, err := frame.Encode()
		if err != nil {
			t.Errorf("error in endoding %s", err)
		}
		reader := bytes.NewReader(buff)
		finalFrame, err := frame.Decoder(reader)
		if err != nil {
			t.Errorf("error in decoding the frame")
		}
		assert.Equal(t, payloadBytes, finalFrame.Payload)
	})
}
