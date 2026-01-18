package main

import (
	"fmt"
	"log"

	"github.com/vjranagit/argo-workflows/pkg/workflow"
)

// This example demonstrates creating a DAG with a diamond dependency pattern.
// Unlike Hera's >> operator, we use explicit dependency methods.
//
// Workflow structure:
//       A
//      / \
//     B   C
//      \ /
//       D
func main() {
	// Create echo template
	echoTemplate := workflow.ScriptTemplate(
		"echo",
		workflow.WithScriptImage("alpine:3.18"),
		workflow.WithScriptCommand("sh"),
		workflow.WithSource(`echo "{{inputs.parameters.message}}"`),
	)

	// Add inputs to the template
	echoTemplate.Inputs = workflow.NewInputs().AddParameter(workflow.Parameter{
		Name: "message",
	})

	// Build the DAG
	dagBuilder := workflow.NewDAG("diamond").
		Task("A", "echo", workflow.WithArguments(
			workflow.NewArguments().AddParameter(workflow.Parameter{
				Name:  "message",
				Value: "A",
			}),
		)).
		Task("B", "echo",
			workflow.WithDependencies("A"),
			workflow.WithArguments(
				workflow.NewArguments().AddParameter(workflow.Parameter{
					Name:  "message",
					Value: "B",
				}),
			),
		).
		Task("C", "echo",
			workflow.WithDependencies("A"),
			workflow.WithArguments(
				workflow.NewArguments().AddParameter(workflow.Parameter{
					Name:  "message",
					Value: "C",
				}),
			),
		).
		Task("D", "echo",
			workflow.WithDependencies("B", "C"),
			workflow.WithArguments(
				workflow.NewArguments().AddParameter(workflow.Parameter{
					Name:  "message",
					Value: "D",
				}),
			),
		)

	// Build the workflow
	wf, err := workflow.New("dag-diamond").
		WithGenerateName("dag-diamond-").
		WithNamespace("argo").
		WithEntrypoint("diamond").
		WithTemplate(echoTemplate).
		WithTemplate(dagBuilder.Build()).
		Build()

	if err != nil {
		log.Fatalf("Build workflow: %v", err)
	}

	fmt.Printf("DAG workflow created: %s\n", wf.Name)
	fmt.Printf("DAG tasks: %d\n", len(wf.Spec.Templates[1].DAG.Tasks))

	// Validate dependencies
	graph := workflow.NewDependencyGraph(wf.Spec.Templates[1].DAG.Tasks)
	if err := graph.Validate(); err != nil {
		log.Fatalf("Invalid DAG: %v", err)
	}

	// Get topological sort
	order, err := graph.TopologicalSort()
	if err != nil {
		log.Fatalf("Topological sort: %v", err)
	}

	fmt.Printf("Execution order: %v\n", order)
	fmt.Println("DAG Diamond workflow created successfully!")
}
