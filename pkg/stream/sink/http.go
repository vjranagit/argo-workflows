package sink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vjranagit/argo-workflows/pkg/stream"
)

// HTTPSink sends data to an HTTP endpoint.
// Unlike Dataflow's Kafka sink, this is a simple HTTP POST.
type HTTPSink[T any] struct {
	url       string
	client    *http.Client
	marshaler func(T) ([]byte, error)
	method    string
	headers   map[string]string
}

// NewHTTP creates a new HTTP sink that POSTs data.
func NewHTTP[T any](url string) *HTTPSink[T] {
	return &HTTPSink[T]{
		url:    url,
		method: "POST",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		marshaler: func(v T) ([]byte, error) {
			return json.Marshal(v)
		},
		headers: map[string]string{
			"Content-Type": "application/json",
		},
	}
}

// WithMethod sets the HTTP method (default: POST).
func (h *HTTPSink[T]) WithMethod(method string) *HTTPSink[T] {
	h.method = method
	return h
}

// WithHeader adds a custom HTTP header.
func (h *HTTPSink[T]) WithHeader(key, value string) *HTTPSink[T] {
	h.headers[key] = value
	return h
}

// WithMarshaler sets a custom marshaler function.
func (h *HTTPSink[T]) WithMarshaler(fn func(T) ([]byte, error)) *HTTPSink[T] {
	h.marshaler = fn
	return h
}

// WithHTTPClient sets a custom HTTP client.
func (h *HTTPSink[T]) WithHTTPClient(client *http.Client) *HTTPSink[T] {
	h.client = client
	return h
}

// Write sends a message to the HTTP endpoint.
func (h *HTTPSink[T]) Write(ctx context.Context, msg stream.Message[T]) error {
	data, err := h.marshaler(msg.Value)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, h.method, h.url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	for key, value := range h.headers {
		req.Header.Set(key, value)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// Close closes the HTTP sink.
func (h *HTTPSink[T]) Close() error {
	if h.client != nil {
		h.client.CloseIdleConnections()
	}
	return nil
}
