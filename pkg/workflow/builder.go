package workflow

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Client interface is forward declared to avoid circular import.
// The actual implementation is in the client package.
type Client interface {
	CreateWorkflow(ctx context.Context, wf *Workflow) (*WorkflowStatus, error)
}

// Builder provides a fluent API for constructing Argo Workflows.
// Unlike Hera's Python decorator and context manager approach, this uses
// explicit method chaining for type-safe workflow construction.
type Builder struct {
	name               string
	namespace          string
	generateName       string
	serviceAccountName string
	entrypoint         string
	templates          []Template
	arguments          *Arguments
	labels             map[string]string
	annotations        map[string]string
}

// New creates a new workflow builder with the given name.
func New(name string) *Builder {
	return &Builder{
		name:        name,
		templates:   make([]Template, 0),
		labels:      make(map[string]string),
		annotations: make(map[string]string),
	}
}

// WithGenerateName sets the generateName field for dynamic naming.
func (b *Builder) WithGenerateName(prefix string) *Builder {
	b.generateName = prefix
	return b
}

// WithNamespace sets the namespace for the workflow.
func (b *Builder) WithNamespace(ns string) *Builder {
	b.namespace = ns
	return b
}

// WithServiceAccount sets the service account for workflow pods.
func (b *Builder) WithServiceAccount(sa string) *Builder {
	b.serviceAccountName = sa
	return b
}

// WithEntrypoint sets the entrypoint template.
func (b *Builder) WithEntrypoint(name string) *Builder {
	b.entrypoint = name
	return b
}

// WithTemplate adds a template to the workflow.
func (b *Builder) WithTemplate(t Template) *Builder {
	b.templates = append(b.templates, t)
	return b
}

// WithArguments sets workflow-level arguments.
func (b *Builder) WithArguments(args *Arguments) *Builder {
	b.arguments = args
	return b
}

// WithLabel adds a label to the workflow.
func (b *Builder) WithLabel(key, value string) *Builder {
	b.labels[key] = value
	return b
}

// WithAnnotation adds an annotation to the workflow.
func (b *Builder) WithAnnotation(key, value string) *Builder {
	b.annotations[key] = value
	return b
}

// Build constructs the final Workflow object.
// This method validates the workflow configuration and returns an error
// if any required fields are missing or invalid.
func (b *Builder) Build() (*Workflow, error) {
	if b.entrypoint == "" {
		return nil, fmt.Errorf("entrypoint is required")
	}

	if len(b.templates) == 0 {
		return nil, fmt.Errorf("at least one template is required")
	}

	// Validate entrypoint exists
	found := false
	for _, t := range b.templates {
		if t.Name == b.entrypoint {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("entrypoint template %q not found", b.entrypoint)
	}

	wf := &Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:         b.name,
			GenerateName: b.generateName,
			Namespace:    b.namespace,
			Labels:       b.labels,
			Annotations:  b.annotations,
		},
		Spec: WorkflowSpec{
			Entrypoint:         b.entrypoint,
			Templates:          b.templates,
			Arguments:          b.arguments,
			ServiceAccountName: b.serviceAccountName,
		},
	}

	return wf, nil
}

// Submit builds and submits the workflow to an Argo server.
// This is different from Hera's w.create() - it uses Go's context
// for cancellation and proper error handling.
func (b *Builder) Submit(ctx context.Context, client Client) (*WorkflowStatus, error) {
	wf, err := b.Build()
	if err != nil {
		return nil, fmt.Errorf("build workflow: %w", err)
	}

	return client.CreateWorkflow(ctx, wf)
}
