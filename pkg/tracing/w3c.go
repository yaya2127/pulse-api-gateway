package tracing

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	HeaderTraceParent = "traceparent"
	HeaderTraceState  = "tracestate"
)

var (
	ErrInvalidTraceParent = errors.New("invalid W3C traceparent header format")
)

func ParseTraceParent(header string) (*TraceContext, error) {
	parts := strings.Split(header, "-")
	if len(parts) != 4 || parts[0] != "00" {
		return nil, ErrInvalidTraceParent
	}

	traceBytes, err := hex.DecodeString(parts[1])
	if err != nil || len(traceBytes) != 16 {
		return nil, ErrInvalidTraceParent
	}

	spanBytes, err := hex.DecodeString(parts[2])
	if err != nil || len(spanBytes) != 8 {
		return nil, ErrInvalidTraceParent
	}

	var tid TraceID
	var sid SpanID
	copy(tid[:], traceBytes)
	copy(sid[:], spanBytes)

	sampled := parts[3] == "01"

	return &TraceContext{
		TraceID:   tid,
		SpanID:    sid,
		Sampled:   sampled,
		StartTime: time.Now(),
	}, nil
}

func FormatTraceParent(tc *TraceContext) string {
	flags := "00"
	if tc.Sampled {
		flags = "01"
	}
	return fmt.Sprintf("00-%s-%s-%s", tc.TraceID.String(), tc.SpanID.String(), flags)
}
