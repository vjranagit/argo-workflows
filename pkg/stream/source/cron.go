package source

import (
	"context"
	"fmt"
	"time"

	"github.com/vjranagit/argo-workflows/pkg/stream"
)

// CronSource generates messages on a cron-like schedule.
// Unlike Dataflow's CRD-based cron, this is an in-process implementation.
type CronSource[T any] struct {
	interval time.Duration
	generator func() T
	out      chan stream.Message[T]
}

// NewCron creates a new cron source that generates messages at the given interval.
func NewCron[T any](interval time.Duration, generator func() T) *CronSource[T] {
	return &CronSource[T]{
		interval:  interval,
		generator: generator,
		out:       make(chan stream.Message[T], 10),
	}
}

// Stream starts generating messages on the schedule.
func (c *CronSource[T]) Stream(ctx context.Context) (<-chan stream.Message[T], error) {
	if c.interval <= 0 {
		return nil, fmt.Errorf("interval must be positive")
	}

	go func() {
		defer close(c.out)

		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				msg := stream.Message[T]{
					Key:       fmt.Sprintf("cron-%d", time.Now().Unix()),
					Value:     c.generator(),
					Timestamp: time.Now().Unix(),
					Metadata:  make(map[string]string),
				}
				select {
				case c.out <- msg:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return c.out, nil
}

// Close closes the cron source.
func (c *CronSource[T]) Close() error {
	return nil
}
