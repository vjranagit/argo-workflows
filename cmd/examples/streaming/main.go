package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/vjranagit/argo-workflows/pkg/stream"
	"github.com/vjranagit/argo-workflows/pkg/stream/sink"
	"github.com/vjranagit/argo-workflows/pkg/stream/source"
)

// This example demonstrates the streaming pipeline feature.
// Unlike Dataflow's CRD-based approach, this is an in-process library.
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a cron source that generates numbers every second
	cronSource := source.NewCron(1*time.Second, func() int {
		return int(time.Now().Unix() % 100)
	})

	// Create a log sink
	logSink := sink.NewLog[int](true)

	// Build and run pipeline
	pipeline := stream.New("number-pipeline", cronSource).
		Filter(func(n int) bool {
			return n%2 == 0 // Only even numbers
		}).
		Map(func(n int) int {
			return n * 2 // Double the value
		}).
		To(logSink)

	fmt.Println("Starting streaming pipeline...")
	fmt.Println("Generating numbers every second, filtering evens, and doubling them")
	fmt.Println("Press Ctrl+C to stop")

	if err := pipeline.Run(ctx); err != nil {
		if err != context.DeadlineExceeded {
			log.Fatalf("Pipeline error: %v", err)
		}
	}

	fmt.Println("\nPipeline completed successfully!")
}
