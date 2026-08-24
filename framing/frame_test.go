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

func TestMaxPayloadExceed(t *testing.T) {
	t.Run("testing for payload more than maximum allowed", func(t *testing.T) {
		task := []byte{1}
		payload := []byte("hello there")
		length := 1024 * 1024
		buf := make([]byte, 1+4+length)
		buf[0] = task[0]
		binary.BigEndian.PutUint32(buf[1:5], uint32(length))
		for i := range len(payload) {
			buf[5+i] = payload[i]
		}
		reader := bytes.NewReader(buf)
		frame := Frame{}
		_, err := frame.Decoder(reader)
		assert.Equal(t, err, errPayloadTooLarge)
	})
}

type TimeRead interface {
	Read(p []byte) (n int, err error)
}
type DelayRead struct {
	reader *bytes.Reader
}

func (d DelayRead) Read(p []byte) (n int, err error) {
	n = 0
	for i := range len(p) {
		num, err := d.reader.Read(p[i : i+1])
		if err != nil {
			break
		}
		n = num + n
	}
	return n, err
}

func TestOnePayloadATime(t *testing.T) {
	t.Run("testing seding payload at a time delay", func(t *testing.T) {
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
		delayread := DelayRead{
			reader,
		}
		frame := Frame{}
		frame.Decoder(delayread)
	})
}

type ExitMidWay struct {
	reader *bytes.Reader
}

func (e ExitMidWay) Read(p []byte) (int, error) {
	num, err := e.reader.Read(p)
	return num, err
}

func TestExitMidWay(t *testing.T) {
	t.Run("testing when the reader gets cut off midway", func(t *testing.T) {
		task := []byte{1}
		payload := []byte("hello there")
		len := len(payload)
		buf := make([]byte, 1+4+len)
		buf[0] = task[0]
		binary.BigEndian.PutUint32(buf[1:5], uint32(len))
		for i := range len {
			buf[5+i] = payload[i]
		}
		reader := bytes.NewReader(buf[:5])
		exitMidWay := ExitMidWay{
			reader,
		}
		frame := Frame{}
		_, err := frame.Decoder(exitMidWay)
		assert.Equal(t, err, errReadPayload)
	})
}
