package tracing

import (
	"crypto/rand"
	"io"
)

type IDGenerator interface {
	NewTraceID() TraceID
	NewSpanID() SpanID
}

type RandomIDGenerator struct{}

func NewRandomIDGenerator() *RandomIDGenerator {
	return &RandomIDGenerator{}
}

func (g *RandomIDGenerator) NewTraceID() TraceID {
	var tid TraceID
	if _, err := io.ReadFull(rand.Reader, tid[:]); err != nil {
		for i := range tid {
			tid[i] = byte(i + 1)
		}
	}
	return tid
}

func (g *RandomIDGenerator) NewSpanID() SpanID {
	var sid SpanID
	if _, err := io.ReadFull(rand.Reader, sid[:]); err != nil {
		for i := range sid {
			sid[i] = byte(i + 1)
		}
	}
	return sid
}
