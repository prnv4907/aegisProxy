package framing

import (
	"encoding/binary"
	"errors"
	"io"
)

const maxFrameSize = 64 * 1024

var (
	errPayloadTooLarge   = errors.New("payload is greater than maximum payload size")
	errReadType          = errors.New("error reading Type")
	errReadPayloadLength = errors.New("error reading payloadLength")
	errReadPayload       = errors.New("error reading payload")
)

type Frame struct {
	Type    byte
	Payload []byte
}

// encode needs to create []byte of 1+4+len(payload)
func (f Frame) Encode() ([]byte, error) {
	if len(f.Payload) > maxFrameSize {
		return nil, errPayloadTooLarge
	}
	buff := make([]byte, 1+4+len(f.Payload)) // Using 5 bytes:1 byte for type,  4 for length of payload
	buff[0] = f.Type
	binary.BigEndian.PutUint32(buff[1:5], uint32(len(f.Payload)))
	copy(buff[5:], f.Payload)
	return buff, nil
}

func (f Frame) Decoder(r io.Reader) (Frame, error) {
	Type := make([]byte, 1)
	_, err := io.ReadFull(r, Type)
	if err != nil {
		if err == io.EOF {
			return Frame{}, io.EOF
		}
		return Frame{}, errReadType
	}
	payloadByte := make([]byte, 4)
	_, err = io.ReadFull(r, payloadByte)
	if err != nil {
		return Frame{}, errReadPayloadLength
	}
	payloadLength := binary.BigEndian.Uint32(payloadByte)

	if int(payloadLength) > maxFrameSize {
		return Frame{}, errPayloadTooLarge
	}
	payload := make([]byte, payloadLength)
	_, err = io.ReadFull(r, payload)
	if err != nil {
		return Frame{}, errReadPayload
	}
	return Frame{Type[0], payload}, nil
}
