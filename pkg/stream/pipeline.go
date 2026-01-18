package stream

import (
	"context"
	"fmt"
	"sync"
)

// Pipeline represents a streaming data pipeline.
// Unlike Dataflow's CRD-based approach, this is an in-process library
// using Go channels for message passing.
type Pipeline[T any] struct {
	name      string
	source    Source[T]
	operators []Operator[T, T]
	sink      Sink[T]
	bufferCap int
	errChan   chan error
}

// New creates a new pipeline with the given name and source.
func New[T any](name string, source Source[T]) *Pipeline[T] {
	return &Pipeline[T]{
		name:      name,
		source:    source,
		operators: make([]Operator[T, T], 0),
		bufferCap: 100,
		errChan:   make(chan error, 10),
	}
}

// WithBufferSize sets the buffer capacity for pipeline channels.
func (p *Pipeline[T]) WithBufferSize(size int) *Pipeline[T] {
	p.bufferCap = size
	return p
}

// Map applies a transformation to each message.
// Unlike Dataflow's expression language, we use Go functions.
func (p *Pipeline[T]) Map(fn func(T) T) *Pipeline[T] {
	p.operators = append(p.operators, &MapOperator[T]{fn: fn})
	return p
}

// Filter removes messages that don't match the predicate.
func (p *Pipeline[T]) Filter(fn func(T) bool) *Pipeline[T] {
	p.operators = append(p.operators, &FilterOperator[T]{fn: fn})
	return p
}

// To sets the sink for the pipeline.
func (p *Pipeline[T]) To(sink Sink[T]) *Pipeline[T] {
	p.sink = sink
	return p
}

// Run executes the pipeline until the context is cancelled or an error occurs.
// This is different from Dataflow's Kubernetes controller approach -
// it's a single-process streaming execution.
func (p *Pipeline[T]) Run(ctx context.Context) error {
	if p.sink == nil {
		return fmt.Errorf("sink is required")
	}

	// Start the source
	sourceChan, err := p.source.Stream(ctx)
	if err != nil {
		return fmt.Errorf("start source: %w", err)
	}

	// Create a chain of channels
	current := sourceChan

	// Apply operators in sequence
	for _, op := range p.operators {
		next, err := op.Process(ctx, current)
		if err != nil {
			return fmt.Errorf("process operator: %w", err)
		}
		current = next
	}

	// Write to sink
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for msg := range current {
			if err := p.sink.Write(ctx, msg); err != nil {
				select {
				case p.errChan <- fmt.Errorf("sink write: %w", err):
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// Wait for completion or context cancellation
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if err := p.sink.Close(); err != nil {
			return fmt.Errorf("close sink: %w", err)
		}
		if err := p.source.Close(); err != nil {
			return fmt.Errorf("close source: %w", err)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case err := <-p.errChan:
		return err
	}
}

// Message represents a single message in the pipeline.
// Generic type parameter allows type-safe message handling.
type Message[T any] struct {
	Key       string
	Value     T
	Timestamp int64
	Metadata  map[string]string
}

// Source defines an interface for pipeline data sources.
// Unlike Dataflow's sidecar pattern, sources are direct Go implementations.
type Source[T any] interface {
	Stream(ctx context.Context) (<-chan Message[T], error)
	Close() error
}

// Sink defines an interface for pipeline data sinks.
type Sink[T any] interface {
	Write(ctx context.Context, msg Message[T]) error
	Close() error
}

// Operator defines an interface for message transformation.
// Uses Go generics for type-safe transformations.
type Operator[In, Out any] interface {
	Process(ctx context.Context, in <-chan Message[In]) (<-chan Message[Out], error)
}

// MapOperator implements a map transformation.
type MapOperator[T any] struct {
	fn func(T) T
}

// Process applies the map function to each message.
func (m *MapOperator[T]) Process(ctx context.Context, in <-chan Message[T]) (<-chan Message[T], error) {
	out := make(chan Message[T], 100)

	go func() {
		defer close(out)
		for msg := range in {
			select {
			case <-ctx.Done():
				return
			default:
				msg.Value = m.fn(msg.Value)
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

// FilterOperator implements a filter transformation.
type FilterOperator[T any] struct {
	fn func(T) bool
}

// Process filters messages based on the predicate.
func (f *FilterOperator[T]) Process(ctx context.Context, in <-chan Message[T]) (<-chan Message[T], error) {
	out := make(chan Message[T], 100)

	go func() {
		defer close(out)
		for msg := range in {
			select {
			case <-ctx.Done():
				return
			default:
				if f.fn(msg.Value) {
					select {
					case out <- msg:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return out, nil
}
