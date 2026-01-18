package source

import (
	"context"

	"github.com/vjranagit/argo-workflows/pkg/stream"
)

// ChannelSource reads messages from a Go channel.
// This is useful for testing and in-memory pipelines.
type ChannelSource[T any] struct {
	in chan T
}

// NewChannel creates a new channel source.
func NewChannel[T any](in chan T) *ChannelSource[T] {
	return &ChannelSource[T]{in: in}
}

// Stream converts raw channel messages to pipeline messages.
func (c *ChannelSource[T]) Stream(ctx context.Context) (<-chan stream.Message[T], error) {
	out := make(chan stream.Message[T], 100)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case val, ok := <-c.in:
				if !ok {
					return
				}
				msg := stream.Message[T]{
					Value:    val,
					Metadata: make(map[string]string),
				}
				select {
				case out <- msg:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, nil
}

// Close closes the channel source.
func (c *ChannelSource[T]) Close() error {
	return nil
}
