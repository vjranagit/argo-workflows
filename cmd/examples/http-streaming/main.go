package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/vjranagit/argo-workflows/pkg/stream"
	"github.com/vjranagit/argo-workflows/pkg/stream/sink"
	"github.com/vjranagit/argo-workflows/pkg/stream/source"
)

// APIResponse represents a sample API response structure.
type APIResponse struct {
	Status string `json:"status"`
	Data   int    `json:"data"`
}

// This example demonstrates HTTP source and sink for data integration.
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Example 1: HTTP Source - Poll an API endpoint
	fmt.Println("=== HTTP Source Example ===")
	
	parser := func(data []byte) (APIResponse, error) {
		var resp APIResponse
		err := json.Unmarshal(data, &resp)
		return resp, err
	}

	// Poll API every 5 seconds
	httpSource := source.NewHTTP(
		"https://api.example.com/metrics",
		5*time.Second,
		parser,
	)

	pipeline1 := stream.New("api-poller", httpSource).
		Filter(func(resp APIResponse) bool {
			return resp.Status == "ok" // Only process successful responses
		}).
		Map(func(resp APIResponse) APIResponse {
			resp.Data = resp.Data * 2 // Transform the data
			return resp
		}).
		To(sink.NewLog[APIResponse](true))

	fmt.Println("HTTP source pipeline configured (would poll https://api.example.com/metrics)")

	// Example 2: HTTP Sink - Send data to an endpoint
	fmt.Println("\n=== HTTP Sink Example ===")

	// Generate data and send via HTTP POST
	generator := func() int {
		return int(time.Now().Unix())
	}

	httpSink := sink.NewHTTP[int]("https://api.example.com/ingest").
		WithMethod("POST").
		WithHeader("Authorization", "Bearer token123").
		WithHeader("X-Source", "argo-workflows")

	pipeline2 := stream.New("http-sender",
		source.NewCron(3*time.Second, generator)).
		Filter(func(n int) bool {
			return n%2 == 0 // Only send even timestamps
		}).
		To(httpSink)

	fmt.Println("HTTP sink pipeline configured (would POST to https://api.example.com/ingest)")

	// Example 3: HTTP to HTTP - API proxy/transformer
	fmt.Println("\n=== HTTP Source → HTTP Sink Pipeline ===")

	pipeline3 := stream.New("api-proxy", httpSource).
		Map(func(resp APIResponse) APIResponse {
			// Transform data before forwarding
			resp.Data = resp.Data + 100
			return resp
		}).
		To(sink.NewHTTP[APIResponse]("https://downstream.example.com/data"))

	fmt.Println("HTTP proxy pipeline configured")
	fmt.Println("  Source: https://api.example.com/metrics")
	fmt.Println("  Sink:   https://downstream.example.com/data")

	fmt.Println("\n✓ HTTP streaming pipelines ready")
	fmt.Println("  (Examples use mock URLs - replace with real endpoints to run)")

	_ = pipeline1
	_ = pipeline2
	_ = pipeline3
}
