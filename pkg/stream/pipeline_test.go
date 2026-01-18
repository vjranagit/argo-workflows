package stream

import (
	"context"
	"testing"
	"time"
)

func TestPipelineBasic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create a simple in-memory source
	in := make(chan int, 10)
	go func() {
		for i := 1; i <= 5; i++ {
			in <- i
		}
		close(in)
	}()

	// Track results
	results := make([]int, 0)
	resultChan := make(chan int, 10)

	// Mock sink
	mockSink := &mockSink[int]{
		writeFn: func(msg Message[int]) error {
			resultChan <- msg.Value
			return nil
		},
	}

	// Build pipeline
	pipeline := New("test", &mockSource[int]{in: in}).
		Filter(func(n int) bool { return n%2 == 0 }).
		Map(func(n int) int { return n * 2 }).
		To(mockSink)

	// Run in goroutine
	done := make(chan error)
	go func() {
		done <- pipeline.Run(ctx)
	}()

	// Collect results
	timeout := time.After(1 * time.Second)
collectLoop:
	for {
		select {
		case val := <-resultChan:
			results = append(results, val)
		case <-timeout:
			break collectLoop
		case err := <-done:
			if err != nil && err != context.Canceled {
				t.Errorf("Pipeline error: %v", err)
			}
			break collectLoop
		}
	}

	// Verify results: filter evens (2, 4) and double them (4, 8)
	expected := []int{4, 8}
	if len(results) != len(expected) {
		t.Errorf("Expected %d results, got %d", len(expected), len(results))
	}

	for i, exp := range expected {
		if i < len(results) && results[i] != exp {
			t.Errorf("Result[%d]: expected %d, got %d", i, exp, results[i])
		}
	}
}

// Mock source for testing
type mockSource[T any] struct {
	in chan T
}

func (m *mockSource[T]) Stream(ctx context.Context) (<-chan Message[T], error) {
	out := make(chan Message[T], 100)
	go func() {
		defer close(out)
		for val := range m.in {
			select {
			case out <- Message[T]{Value: val}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

func (m *mockSource[T]) Close() error {
	return nil
}

// Mock sink for testing
type mockSink[T any] struct {
	writeFn func(Message[T]) error
}

func (m *mockSink[T]) Write(ctx context.Context, msg Message[T]) error {
	return m.writeFn(msg)
}

func (m *mockSink[T]) Close() error {
	return nil
}
