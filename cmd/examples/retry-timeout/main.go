package main

import (
	"context"
	"fmt"
	"log"

	"github.com/vjranagit/argo-workflows/pkg/workflow"
)

// This example demonstrates retry and timeout policies for resilient workflows.
func main() {
	ctx := context.Background()

	// Create a flaky task template with retry logic
	flakyTask := workflow.ContainerTemplate(
		"flaky-task",
		workflow.WithImage("alpine:latest"),
		workflow.WithCommand("sh", "-c", "if [ $RANDOM -lt 16384 ]; then exit 1; else echo 'Success!'; fi"),
		// Retry up to 5 times with exponential backoff
		workflow.WithRetryStrategy(5, workflow.RetryPolicyOnFailure),
		workflow.WithRetryBackoff("10s", 2, "5m"),
		// Overall timeout of 30 minutes
		workflow.WithTimeout("30m"),
	)

	// Create a long-running task with just a timeout
	longTask := workflow.ContainerTemplate(
		"long-task",
		workflow.WithImage("alpine:latest"),
		workflow.WithCommand("sleep", "300"),
		workflow.WithTimeout("5m"), // Will timeout after 5 minutes
	)

	// Build workflow with retry-enabled tasks
	wf, err := workflow.New("retry-example").
		WithGenerateName("retry-example-").
		WithNamespace("argo").
		WithEntrypoint("main").
		WithTemplate(flakyTask).
		WithTemplate(longTask).
		WithTemplate(workflow.NewDAG("main").
			Task("flaky", "flaky-task").
			Task("long", "long-task", workflow.WithDependencies("flaky")).
			Build()).
		Build()

	if err != nil {
		log.Fatalf("Build workflow: %v", err)
	}

	// Export to YAML for GitOps
	yamlData, err := wf.ToYAML()
	if err != nil {
		log.Fatalf("Export YAML: %v", err)
	}

	fmt.Println("Workflow with retry and timeout policies:")
	fmt.Println(string(yamlData))

	// Save to file
	if err := wf.ToYAMLFile("retry-workflow.yaml"); err != nil {
		log.Fatalf("Save YAML: %v", err)
	}

	fmt.Println("\n✓ Saved to retry-workflow.yaml")
}
