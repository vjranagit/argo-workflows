package workflow

import (
	"os"
	"strings"
	"testing"
)

func TestToYAML(t *testing.T) {
	wf, err := New("test-workflow").
		WithNamespace("default").
		WithEntrypoint("main").
		WithTemplate(ContainerTemplate(
			"main",
			WithImage("alpine:latest"),
			WithCommand("echo", "hello"),
		)).
		Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	yaml, err := wf.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	// Verify YAML contains expected fields
	yamlStr := string(yaml)
	if !strings.Contains(yamlStr, "apiVersion: argoproj.io/v1alpha1") {
		t.Error("YAML missing apiVersion")
	}
	if !strings.Contains(yamlStr, "kind: Workflow") {
		t.Error("YAML missing kind")
	}
	if !strings.Contains(yamlStr, "name: test-workflow") {
		t.Error("YAML missing workflow name")
	}
	if !strings.Contains(yamlStr, "entrypoint: main") {
		t.Error("YAML missing entrypoint")
	}
}

func TestFromYAML(t *testing.T) {
	yamlData := []byte(`
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  name: test-workflow
  namespace: default
spec:
  entrypoint: main
  templates:
  - name: main
    container:
      image: alpine:latest
      command: [echo, hello]
`)

	wf, err := FromYAML(yamlData)
	if err != nil {
		t.Fatalf("FromYAML() error = %v", err)
	}

	if wf.Name != "test-workflow" {
		t.Errorf("Name = %v, want test-workflow", wf.Name)
	}

	if wf.Namespace != "default" {
		t.Errorf("Namespace = %v, want default", wf.Namespace)
	}

	if wf.Spec.Entrypoint != "main" {
		t.Errorf("Entrypoint = %v, want main", wf.Spec.Entrypoint)
	}

	if len(wf.Spec.Templates) != 1 {
		t.Errorf("Templates count = %v, want 1", len(wf.Spec.Templates))
	}
}

func TestYAMLRoundTrip(t *testing.T) {
	// Create a workflow
	original, err := New("roundtrip-test").
		WithNamespace("test-ns").
		WithEntrypoint("main").
		WithTemplate(ContainerTemplate(
			"main",
			WithImage("alpine:latest"),
			WithCommand("echo", "test"),
			WithRetryStrategy(3, RetryPolicyOnFailure),
			WithTimeout("5m"),
		)).
		Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Convert to YAML
	yamlData, err := original.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	// Parse back from YAML
	parsed, err := FromYAML(yamlData)
	if err != nil {
		t.Fatalf("FromYAML() error = %v", err)
	}

	// Verify key fields match
	if parsed.Name != original.Name {
		t.Errorf("Name = %v, want %v", parsed.Name, original.Name)
	}

	if parsed.Namespace != original.Namespace {
		t.Errorf("Namespace = %v, want %v", parsed.Namespace, original.Namespace)
	}

	if parsed.Spec.Entrypoint != original.Spec.Entrypoint {
		t.Errorf("Entrypoint = %v, want %v", parsed.Spec.Entrypoint, original.Spec.Entrypoint)
	}

	if len(parsed.Spec.Templates) != len(original.Spec.Templates) {
		t.Errorf("Templates count = %v, want %v", len(parsed.Spec.Templates), len(original.Spec.Templates))
	}

	// Verify retry strategy was preserved
	if parsed.Spec.Templates[0].RetryStrategy == nil {
		t.Error("RetryStrategy should not be nil after roundtrip")
	}

	// Verify timeout was preserved
	if parsed.Spec.Templates[0].Timeout == nil {
		t.Error("Timeout should not be nil after roundtrip")
	}
}

func TestToYAMLFile(t *testing.T) {
	wf, err := New("file-test").
		WithNamespace("default").
		WithEntrypoint("main").
		WithTemplate(ContainerTemplate(
			"main",
			WithImage("alpine:latest"),
		)).
		Build()

	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Write to temp file
	tmpfile := "/tmp/test-workflow.yaml"
	defer os.Remove(tmpfile)

	if err := wf.ToYAMLFile(tmpfile); err != nil {
		t.Fatalf("ToYAMLFile() error = %v", err)
	}

	// Read back
	loaded, err := FromYAMLFile(tmpfile)
	if err != nil {
		t.Fatalf("FromYAMLFile() error = %v", err)
	}

	if loaded.Name != wf.Name {
		t.Errorf("Name = %v, want %v", loaded.Name, wf.Name)
	}
}

func TestYAMLBuilder(t *testing.T) {
	yamlData := []byte(`
apiVersion: argoproj.io/v1alpha1
kind: Workflow
metadata:
  name: builder-test
spec:
  entrypoint: main
  templates:
  - name: main
    container:
      image: alpine:latest
`)

	builder, err := NewFromYAML(yamlData)
	if err != nil {
		t.Fatalf("NewFromYAML() error = %v", err)
	}

	wf := builder.Workflow()
	if wf.Name != "builder-test" {
		t.Errorf("Name = %v, want builder-test", wf.Name)
	}

	// Test String() method
	yamlStr, err := builder.String()
	if err != nil {
		t.Fatalf("String() error = %v", err)
	}

	if !strings.Contains(yamlStr, "builder-test") {
		t.Error("String output missing workflow name")
	}
}

func TestInvalidYAML(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		wantErr  bool
	}{
		{
			name:    "invalid kind",
			yaml:    `kind: Pod`,
			wantErr: true,
		},
		{
			name:    "malformed yaml",
			yaml:    `{invalid yaml [}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FromYAML([]byte(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("FromYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
