package tracing

import (
	"context"
	"net/http"
	"time"
)

type Middleware struct {
	gen IDGenerator
}

func NewMiddleware(gen IDGenerator) *Middleware {
	if gen == nil {
		gen = NewRandomIDGenerator()
	}
	return &Middleware{gen: gen}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc, err := ParseTraceParent(r.Header.Get(HeaderTraceParent))
		if err != nil {
			tc = &TraceContext{
				TraceID:   m.gen.NewTraceID(),
				SpanID:    m.gen.NewSpanID(),
				Sampled:   true,
				StartTime: time.Now(),
			}
		} else {
			tc.ParentID = tc.SpanID
			tc.SpanID = m.gen.NewSpanID()
		}

		w.Header().Set(HeaderTraceParent, FormatTraceParent(tc))
		ctx := context.WithValue(r.Context(), TraceContextKey, tc)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func FromContext(ctx context.Context) *TraceContext {
	if tc, ok := ctx.Value(TraceContextKey).(*TraceContext); ok {
		return tc
	}
	return nil
}
