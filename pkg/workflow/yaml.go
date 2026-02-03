package workflow

import (
	"fmt"
	"io"
	"os"

	"sigs.k8s.io/yaml"
)

// ToYAML serializes a workflow to YAML format.
// This enables GitOps workflows and workflow portability.
func (wf *Workflow) ToYAML() ([]byte, error) {
	// Set API version and kind
	wf.APIVersion = "argoproj.io/v1alpha1"
	wf.Kind = "Workflow"

	data, err := yaml.Marshal(wf)
	if err != nil {
		return nil, fmt.Errorf("marshal workflow: %w", err)
	}

	return data, nil
}

// ToYAMLFile writes a workflow to a YAML file.
func (wf *Workflow) ToYAMLFile(filename string) error {
	data, err := wf.ToYAML()
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// FromYAML deserializes a workflow from YAML.
func FromYAML(data []byte) (*Workflow, error) {
	var wf Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, fmt.Errorf("unmarshal workflow: %w", err)
	}

	// Validate basic fields
	if wf.Kind != "Workflow" {
		return nil, fmt.Errorf("invalid kind: %s (expected Workflow)", wf.Kind)
	}

	return &wf, nil
}

// FromYAMLFile reads a workflow from a YAML file.
func FromYAMLFile(filename string) (*Workflow, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return FromYAML(data)
}

// FromYAMLReader reads a workflow from an io.Reader.
func FromYAMLReader(r io.Reader) (*Workflow, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read data: %w", err)
	}

	return FromYAML(data)
}

// YAMLBuilder provides a fluent API for YAML workflow operations.
type YAMLBuilder struct {
	wf *Workflow
}

// NewFromYAML creates a builder from YAML data.
func NewFromYAML(data []byte) (*YAMLBuilder, error) {
	wf, err := FromYAML(data)
	if err != nil {
		return nil, err
	}

	return &YAMLBuilder{wf: wf}, nil
}

// NewFromYAMLFile creates a builder from a YAML file.
func NewFromYAMLFile(filename string) (*YAMLBuilder, error) {
	wf, err := FromYAMLFile(filename)
	if err != nil {
		return nil, err
	}

	return &YAMLBuilder{wf: wf}, nil
}

// Workflow returns the workflow object.
func (yb *YAMLBuilder) Workflow() *Workflow {
	return yb.wf
}

// Save writes the workflow to a YAML file.
func (yb *YAMLBuilder) Save(filename string) error {
	return yb.wf.ToYAMLFile(filename)
}

// Bytes returns the YAML bytes.
func (yb *YAMLBuilder) Bytes() ([]byte, error) {
	return yb.wf.ToYAML()
}

// String returns the YAML string.
func (yb *YAMLBuilder) String() (string, error) {
	data, err := yb.wf.ToYAML()
	if err != nil {
		return "", err
	}
	return string(data), nil
}
