package wasm

import (
	"context"
	"net/http"
)

type FilterContext struct {
	RequestID string
	RemoteIP  string
	Path      string
}

type PluginRuntime interface {
	Init(config string) error
	OnHttpRequestHeaders(ctx *FilterContext, headers http.Header) error
	OnHttpResponseHeaders(ctx *FilterContext, headers http.Header) error
}

type WASMFilter struct {
	plugins []PluginRuntime
}

func NewWASMFilter() *WASMFilter {
	return &WASMFilter{plugins: make([]PluginRuntime, 0)}
}

func (f *WASMFilter) RegisterPlugin(p PluginRuntime) {
	f.plugins = append(f.plugins, p)
}

func (f *WASMFilter) ExecuteRequestHooks(ctx *FilterContext, h http.Header) error {
	for _, p := range f.plugins {
		if err := p.OnHttpRequestHeaders(ctx, h); err != nil {
			return err
		}
	}
	return nil
}
