package sink

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/vjranagit/argo-workflows/pkg/stream"
)

// LogSink writes messages to standard output as JSON.
// Different from Dataflow's log sink - this uses structured JSON logging.
type LogSink[T any] struct {
	logger *log.Logger
	pretty bool
}

// NewLog creates a new log sink.
func NewLog[T any](pretty bool) *LogSink[T] {
	return &LogSink[T]{
		logger: log.New(os.Stdout, "[PIPELINE] ", log.LstdFlags),
		pretty: pretty,
	}
}

// Write logs the message as JSON.
func (l *LogSink[T]) Write(ctx context.Context, msg stream.Message[T]) error {
	var data []byte
	var err error

	if l.pretty {
		data, err = json.MarshalIndent(msg, "", "  ")
	} else {
		data, err = json.Marshal(msg)
	}

	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	l.logger.Println(string(data))
	return nil
}

// Close closes the log sink.
func (l *LogSink[T]) Close() error {
	return nil
}
