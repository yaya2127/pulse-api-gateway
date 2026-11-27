package grpcweb

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	FrameData     byte = 0x00
	FrameTrailers byte = 0x80
)

type Frame struct {
	Flag    byte
	Length  uint32
	Payload []byte
}

func EncodeFrame(flag byte, payload []byte) []byte {
	buf := make([]byte, 5+len(payload))
	buf[0] = flag
	binary.BigEndian.PutUint32(buf[1:5], uint32(len(payload)))
	copy(buf[5:], payload)
	return buf
}

func DecodeFrame(r io.Reader) (*Frame, error) {
	header := make([]byte, 5)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	flag := header[0]
	length := binary.BigEndian.Uint32(header[1:5])

	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}

	return &Frame{
		Flag:    flag,
		Length:  length,
		Payload: payload,
	}, nil
}
