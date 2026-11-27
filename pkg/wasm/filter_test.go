package wasm

import (
	"net/http"
	"testing"
)

type MockPlugin struct{}

func (m *MockPlugin) Init(config string) error { return nil }
func (m *MockPlugin) OnHttpRequestHeaders(ctx *FilterContext, headers http.Header) error {
	headers.Set("X-Wasm-Injected", "true")
	return nil
}
func (m *MockPlugin) OnHttpResponseHeaders(ctx *FilterContext, headers http.Header) error { return nil }

func TestWASMFilter(t *testing.T) {
	filter := NewWASMFilter()
	filter.RegisterPlugin(&MockPlugin{})

	headers := make(http.Header)
	ctx := &FilterContext{RequestID: "req-1"}

	if err := filter.ExecuteRequestHooks(ctx, headers); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if headers.Get("X-Wasm-Injected") != "true" {
		t.Fatal("expected X-Wasm-Injected header to be set by WASM plugin")
	}
}
