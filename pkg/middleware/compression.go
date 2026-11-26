package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

type CompressionMiddleware struct {
	pool sync.Pool
}

func NewCompressionMiddleware() *CompressionMiddleware {
	return &CompressionMiddleware{
		pool: sync.Pool{
			New: func() any {
				w, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
				return w
			},
		},
	}
}
