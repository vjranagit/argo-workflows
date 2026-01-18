package workflow

import (
	"testing"
)

func TestBuilderBasic(t *testing.T) {
	template := ContainerTemplate(
		"test",
		WithImage("alpine:3.18"),
		WithCommand("echo", "hello"),
	)

	wf, err := New("test-workflow").
		WithNamespace("default").
		WithEntrypoint("test").
		WithTemplate(template).
		Build()

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if wf.Name != "test-workflow" {
		t.Errorf("Expected name 'test-workflow', got '%s'", wf.Name)
	}

	if wf.Namespace != "default" {
		t.Errorf("Expected namespace 'default', got '%s'", wf.Namespace)
	}

	if wf.Spec.Entrypoint != "test" {
		t.Errorf("Expected entrypoint 'test', got '%s'", wf.Spec.Entrypoint)
	}

	if len(wf.Spec.Templates) != 1 {
		t.Errorf("Expected 1 template, got %d", len(wf.Spec.Templates))
	}
}

func TestBuilderMissingEntrypoint(t *testing.T) {
	template := ContainerTemplate("test", WithImage("alpine:3.18"))

	_, err := New("test-workflow").
		WithTemplate(template).
		Build()

	if err == nil {
		t.Fatal("Expected error for missing entrypoint")
	}
}

func TestBuilderInvalidEntrypoint(t *testing.T) {
	template := ContainerTemplate("test", WithImage("alpine:3.18"))

	_, err := New("test-workflow").
		WithEntrypoint("nonexistent").
		WithTemplate(template).
		Build()

	if err == nil {
		t.Fatal("Expected error for invalid entrypoint")
	}
}
