package main

import (
	"context"
	"fmt"
	"log"

	"github.com/vjranagit/argo-workflows/pkg/client"
	"github.com/vjranagit/argo-workflows/pkg/workflow"
)

// This example demonstrates creating a simple "hello world" workflow.
// Unlike Hera's Python approach with decorators, this uses Go's builder pattern.
func main() {
	ctx := context.Background()

	// Create a simple container template
	helloTemplate := workflow.ContainerTemplate(
		"whalesay",
		workflow.WithImage("docker/whalesay:latest"),
		workflow.WithCommand("cowsay", "hello world"),
	)

	// Build the workflow
	wf, err := workflow.New("hello-world").
		WithGenerateName("hello-world-").
		WithNamespace("argo").
		WithEntrypoint("whalesay").
		WithTemplate(helloTemplate).
		Build()

	if err != nil {
		log.Fatalf("Build workflow: %v", err)
	}

	// Print the workflow spec
	fmt.Printf("Workflow: %s\n", wf.Name)
	fmt.Printf("Entrypoint: %s\n", wf.Spec.Entrypoint)
	fmt.Printf("Templates: %d\n", len(wf.Spec.Templates))

	// To submit to Argo server (when configured):
	// client := client.NewHTTPClient(client.Config{
	//     BaseURL:   "http://localhost:2746",
	//     Namespace: "argo",
	//     Auth:      client.NewNoAuth(),
	// })
	// status, err := workflow.New("hello-world").
	//     WithTemplate(helloTemplate).
	//     Submit(ctx, client)

	_ = ctx
	_ = client.Client(nil)
	fmt.Println("Hello World workflow created successfully!")
}
