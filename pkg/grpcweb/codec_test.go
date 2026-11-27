package grpcweb

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeFrame(t *testing.T) {
	payload := []byte("grpc-protobuf-serialized-body")
	encoded := EncodeFrame(FrameData, payload)

	if len(encoded) != 5+len(payload) {
		t.Fatalf("unexpected frame length: %d", len(encoded))
	}

	frame, err := DecodeFrame(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("failed to decode frame: %v", err)
	}

	if frame.Flag != FrameData {
		t.Errorf("got flag %x, want %x", frame.Flag, FrameData)
	}
	if !bytes.Equal(frame.Payload, payload) {
		t.Fatalf("payload mismatch: %s != %s", string(frame.Payload), string(payload))
	}
}
