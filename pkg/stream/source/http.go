package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vjranagit/argo-workflows/pkg/stream"
)

// HTTPSource polls an HTTP endpoint for data.
// Unlike Dataflow's webhook approach, this is pull-based polling.
type HTTPSource[T any] struct {
	url      string
	interval time.Duration
	client   *http.Client
	parser   func([]byte) (T, error)
	ch       chan stream.Message[T]
}

// NewHTTP creates a new HTTP polling source.
func NewHTTP[T any](url string, interval time.Duration, parser func([]byte) (T, error)) *HTTPSource[T] {
	if parser == nil {
		// Default JSON parser
		parser = func(data []byte) (T, error) {
			var result T
			err := json.Unmarshal(data, &result)
			return result, err
		}
	}

	return &HTTPSource[T]{
		url:      url,
		interval: interval,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		parser: parser,
	}
}

// WithHTTPClient sets a custom HTTP client.
func (h *HTTPSource[T]) WithHTTPClient(client *http.Client) *HTTPSource[T] {
	h.client = client
	return h
}

// Stream starts polling the HTTP endpoint.
func (h *HTTPSource[T]) Stream(ctx context.Context) (<-chan stream.Message[T], error) {
	h.ch = make(chan stream.Message[T], 10)

	go func() {
		defer close(h.ch)

		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()

		// Poll immediately, then on interval
		h.poll(ctx)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				h.poll(ctx)
			}
		}
	}()

	return h.ch, nil
}

func (h *HTTPSource[T]) poll(ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, "GET", h.url, nil)
	if err != nil {
		return
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	value, err := h.parser(body)
	if err != nil {
		return
	}

	msg := stream.Message[T]{
		Key:       h.url,
		Value:     value,
		Timestamp: time.Now().Unix(),
		Metadata: map[string]string{
			"source": "http",
			"url":    h.url,
			"status": fmt.Sprintf("%d", resp.StatusCode),
		},
	}

	select {
	case h.ch <- msg:
	case <-ctx.Done():
	}
}

// Close closes the HTTP source.
func (h *HTTPSource[T]) Close() error {
	if h.client != nil {
		h.client.CloseIdleConnections()
	}
	return nil
}
