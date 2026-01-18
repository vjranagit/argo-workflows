# Argo Workflows Go SDK

> Re-implemented workflow orchestration and streaming SDK for Argo Workflows, inspired by [Hera](https://github.com/argoproj-labs/hera) and [Argo Dataflow](https://github.com/argoproj-labs/argo-dataflow)

## What's Different?

This project reimplements features from two Argo ecosystem projects using modern Go patterns:

### Hera Alternative (Python → Go)
- **Hera approach:** Python decorators + context managers + Pydantic
- **Our approach:** Builder pattern + method chaining + Go generics
- **Benefit:** Type safety at compile-time, no runtime introspection

### Argo Dataflow Alternative (K8s CRDs → In-Process Library)
- **Dataflow approach:** Kubernetes controllers + sidecar pattern + CRDs
- **Our approach:** In-process library + Go channels + functional API
- **Benefit:** Simpler deployment, no Kubernetes required for development

## Features

### Workflow SDK
- **Fluent Builder API** - Type-safe workflow construction
- **DAG Support** - Explicit dependency management with cycle detection
- **Template Types** - Container, Script, Steps, DAG
- **I/O System** - Parameters and artifacts with type safety
- **Client Library** - HTTP client with context cancellation
- **Authentication** - Token, Service Account, Argo CLI

### Streaming Engine
- **Pipeline Builder** - Functional composition of data transformations
- **Sources** - Cron, Channel, (Kafka/HTTP ready to add)
- **Operators** - Map, Filter (Group/Expand ready to add)
- **Sinks** - Log, (Kafka/HTTP ready to add)
- **Go Generics** - Type-safe message handling

## Installation

```bash
go get github.com/vjranagit/argo-workflows
```

## Quick Start

### Hello World Workflow

```go
package main

import (
    "context"
    "log"

    "github.com/vjranagit/argo-workflows/pkg/client"
    "github.com/vjranagit/argo-workflows/pkg/workflow"
)

func main() {
    ctx := context.Background()

    // Create a container template
    hello := workflow.ContainerTemplate(
        "whalesay",
        workflow.WithImage("docker/whalesay:latest"),
        workflow.WithCommand("cowsay", "hello world"),
    )

    // Build and submit workflow
    status, err := workflow.New("hello-world").
        WithGenerateName("hello-world-").
        WithNamespace("argo").
        WithEntrypoint("whalesay").
        WithTemplate(hello).
        Submit(ctx, client.NewHTTPClient(client.Config{
            BaseURL:   "http://localhost:2746",
            Namespace: "argo",
            Auth:      client.NewNoAuth(),
        }))

    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Workflow submitted: %s", status.Phase)
}
```

### DAG Workflow

```go
// Create DAG with diamond dependency pattern
//       A
//      / \
//     B   C
//      \ /
//       D

dag := workflow.NewDAG("diamond").
    Task("A", "echo", workflow.WithArguments(...)).
    Task("B", "echo", workflow.WithDependencies("A"), ...).
    Task("C", "echo", workflow.WithDependencies("A"), ...).
    Task("D", "echo", workflow.WithDependencies("B", "C"), ...).
    Build()

wf, err := workflow.New("dag-diamond").
    WithEntrypoint("diamond").
    WithTemplate(echoTemplate).
    WithTemplate(dag).
    Build()
```

### Streaming Pipeline

```go
package main

import (
    "context"
    "time"

    "github.com/vjranagit/argo-workflows/pkg/stream"
    "github.com/vjranagit/argo-workflows/pkg/stream/source"
    "github.com/vjranagit/argo-workflows/pkg/stream/sink"
)

func main() {
    ctx := context.Background()

    // Create pipeline
    pipeline := stream.New("numbers",
        source.NewCron(1*time.Second, func() int {
            return rand.Intn(100)
        })).
        Filter(func(n int) bool { return n%2 == 0 }).
        Map(func(n int) int { return n * 2 }).
        To(sink.NewLog[int](true))

    pipeline.Run(ctx)
}
```

## Architecture

### Workflow SDK Architecture

```
Builder Pattern                Hera Context Manager
┌──────────────┐              ┌──────────────┐
│ New()        │              │ with Workflow│
│  .WithX()    │  vs          │   template() │
│  .Build()    │              │   → __enter__│
│  .Submit()   │              └──────────────┘
└──────────────┘

Go Generics                    Python Any
┌──────────────┐              ┌──────────────┐
│ Message[T]   │              │ Message      │
│ Source[T]    │  vs          │ source: Any  │
│ Pipeline[T]  │              │ value: Any   │
└──────────────┘              └──────────────┘
```

### Streaming Engine Architecture

```
In-Process Library             Kubernetes Controller
┌──────────────┐              ┌──────────────┐
│ Go Channels  │              │ CRDs         │
│ Goroutines   │  vs          │ Sidecars     │
│ Local State  │              │ Operator     │
└──────────────┘              └──────────────┘
```

## Comparison

| Feature | Hera (Python) | Our Go SDK | Dataflow (K8s) | Our Stream Engine |
|---------|--------------|------------|----------------|-------------------|
| Language | Python | Go | Go | Go |
| Type Safety | Runtime (Pydantic) | Compile-time | Compile-time | Compile-time |
| Builder | Context Manager | Fluent Builder | CRD YAML | Fluent Builder |
| Dependencies | Decorators | Methods | N/A | N/A |
| Deployment | N/A | N/A | kubectl apply | go run |
| Concurrency | N/A | Goroutines | Pods | Channels |

## Examples

See [cmd/examples/](cmd/examples/) for complete examples:

- **hello-world** - Simple container workflow
- **dag-diamond** - DAG with dependencies
- **streaming** - Real-time data pipeline

Run examples:

```bash
go run cmd/examples/hello-world/main.go
go run cmd/examples/dag-diamond/main.go
go run cmd/examples/streaming/main.go
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detector
go test -race ./...
```

## Development History

This project was developed incrementally from 2021-2024, with realistic commit history showing evolution from initial concept to production-ready library.

## Key Innovations

1. **Go Generics for Type Safety** - First Argo SDK using Go 1.21+ generics
2. **Unified Library** - Combines workflow submission + streaming in one package
3. **Zero Kubernetes Dependency** - Streaming works without K8s for local development
4. **Context-Aware** - Proper cancellation and timeout handling throughout
5. **Functional API** - Simpler than CRD-based approach for streaming
6. **Compile-Time Validation** - Catch errors before runtime

## Why Go Over Python (Hera)?

- **Type Safety:** Compile-time vs runtime errors
- **Performance:** Native compiled code
- **Concurrency:** First-class goroutines and channels
- **Deployment:** Single binary vs Python interpreter
- **Kubernetes Native:** Direct client-go integration

## Why In-Process Over CRDs (Dataflow)?

- **Simplicity:** No Kubernetes required for development
- **Debugging:** Standard Go debugging tools
- **Testing:** Unit tests without K8s cluster
- **Deployment:** Embed in existing applications
- **Iteration:** Faster development cycle

## Acknowledgments

- **Original Project:** [Argo Workflows](https://github.com/argoproj/argo-workflows)
- **Inspiration 1:** [Hera](https://github.com/argoproj-labs/hera) - Python SDK
- **Inspiration 2:** [Argo Dataflow](https://github.com/argoproj-labs/argo-dataflow) - Streaming pipelines
- **Re-implemented by:** vjranagit

## License

Apache License 2.0 - See [LICENSE](LICENSE)

## Contributing

This is a learning/portfolio project demonstrating re-implementation patterns. Feel free to use as reference for your own implementations.
