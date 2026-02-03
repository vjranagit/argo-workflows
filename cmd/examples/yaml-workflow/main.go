package main

import (
	"fmt"
	"log"

	"github.com/vjranagit/argo-workflows/pkg/workflow"
)

// This example demonstrates YAML export/import for GitOps workflows.
func main() {
	// Create a complex workflow programmatically
	echoTemplate := workflow.ScriptTemplate(
		"echo",
		workflow.WithScriptImage("alpine:3.18"),
		workflow.WithSource(`echo "Processing: $MESSAGE"`),
		workflow.WithRetryStrategy(3, workflow.RetryPolicyOnFailure),
	)

	dagTemplate := workflow.NewDAG("pipeline").
		Task("extract", "echo", workflow.WithArguments(
			workflow.NewArguments().AddParameter(workflow.Parameter{
				Name:  "message",
				Value: "Extracting data",
			}),
		)).
		Task("transform", "echo", 
			workflow.WithDependencies("extract"),
			workflow.WithArguments(
				workflow.NewArguments().AddParameter(workflow.Parameter{
					Name:  "message",
					Value: "Transforming data",
				}),
			),
		).
		Task("load", "echo",
			workflow.WithDependencies("transform"),
			workflow.WithArguments(
				workflow.NewArguments().AddParameter(workflow.Parameter{
					Name:  "message",
					Value: "Loading data",
				}),
			),
		).
		Build()

	wf, err := workflow.New("etl-pipeline").
		WithGenerateName("etl-").
		WithNamespace("data-engineering").
		WithLabel("team", "data").
		WithLabel("project", "analytics").
		WithAnnotation("description", "ETL pipeline with retry policies").
		WithEntrypoint("pipeline").
		WithTemplate(echoTemplate).
		WithTemplate(dagTemplate).
		Build()

	if err != nil {
		log.Fatalf("Build workflow: %v", err)
	}

	// Export to YAML
	fmt.Println("=== Exporting Workflow to YAML ===\n")
	yamlData, err := wf.ToYAML()
	if err != nil {
		log.Fatalf("Export YAML: %v", err)
	}
	fmt.Println(string(yamlData))

	// Save to file (for GitOps)
	if err := wf.ToYAMLFile("etl-pipeline.yaml"); err != nil {
		log.Fatalf("Save YAML file: %v", err)
	}
	fmt.Println("\n✓ Saved to etl-pipeline.yaml")

	// Load from YAML
	fmt.Println("\n=== Loading Workflow from YAML ===\n")
	loaded, err := workflow.FromYAMLFile("etl-pipeline.yaml")
	if err != nil {
		log.Fatalf("Load YAML: %v", err)
	}

	fmt.Printf("Loaded workflow: %s\n", loaded.Name)
	fmt.Printf("Namespace: %s\n", loaded.Namespace)
	fmt.Printf("Entrypoint: %s\n", loaded.Spec.Entrypoint)
	fmt.Printf("Templates: %d\n", len(loaded.Spec.Templates))
	fmt.Printf("Labels: %v\n", loaded.Labels)

	// Use YAML builder for fluent operations
	fmt.Println("\n=== Using YAML Builder ===\n")
	builder, err := workflow.NewFromYAMLFile("etl-pipeline.yaml")
	if err != nil {
		log.Fatalf("Create builder: %v", err)
	}

	yamlStr, err := builder.String()
	if err != nil {
		log.Fatalf("Get YAML string: %v", err)
	}

	fmt.Println("Workflow as YAML string (first 200 chars):")
	if len(yamlStr) > 200 {
		fmt.Println(yamlStr[:200] + "...")
	} else {
		fmt.Println(yamlStr)
	}

	fmt.Println("\n✓ YAML export/import demonstration complete")
	fmt.Println("  Use these files in GitOps workflows with ArgoCD!")
}
