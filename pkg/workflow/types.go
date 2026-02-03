package workflow

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Workflow represents an Argo Workflow resource.
// This is our own representation, different from Hera's Pydantic models.
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              WorkflowSpec   `json:"spec"`
	Status            WorkflowStatus `json:"status,omitempty"`
}

// WorkflowSpec defines the desired state of a Workflow.
type WorkflowSpec struct {
	Entrypoint         string      `json:"entrypoint"`
	Templates          []Template  `json:"templates"`
	Arguments          *Arguments  `json:"arguments,omitempty"`
	ServiceAccountName string      `json:"serviceAccountName,omitempty"`
	Parallelism        *int32      `json:"parallelism,omitempty"`
	ActiveDeadline     *int64      `json:"activeDeadlineSeconds,omitempty"`
	TTL                *int32      `json:"ttlSecondsAfterFinished,omitempty"`
}

// Template defines a workflow template.
// Unlike Hera's class-based approach, we use composition with a union-like structure.
type Template struct {
	Name      string         `json:"name"`
	Inputs    *Inputs        `json:"inputs,omitempty"`
	Outputs   *Outputs       `json:"outputs,omitempty"`
	Container *Container     `json:"container,omitempty"`
	Script    *Script        `json:"script,omitempty"`
	DAG       *DAG           `json:"dag,omitempty"`
	Steps          *[][]StepGroup  `json:"steps,omitempty"`
	RetryStrategy  *RetryStrategy  `json:"retryStrategy,omitempty"`
	Timeout        *TimeoutPolicy  `json:"timeout,omitempty"`
}

// Container defines a container template.
type Container struct {
	Name       string        `json:"name,omitempty"`
	Image      string        `json:"image"`
	Command    []string      `json:"command,omitempty"`
	Args       []string      `json:"args,omitempty"`
	Env        []EnvVar      `json:"env,omitempty"`
	Resources  *Resources    `json:"resources,omitempty"`
	WorkingDir string        `json:"workingDir,omitempty"`
}

// Script defines a script template.
// Different from Hera's @script decorator - this is explicit.
type Script struct {
	Image      string     `json:"image"`
	Command    []string   `json:"command,omitempty"`
	Source     string     `json:"source"`
	Env        []EnvVar   `json:"env,omitempty"`
	Resources  *Resources `json:"resources,omitempty"`
	WorkingDir string     `json:"workingDir,omitempty"`
}

// DAG defines a directed acyclic graph template.
// Unlike Hera's >> operator approach, we use explicit task lists.
type DAG struct {
	Tasks []DAGTask `json:"tasks"`
}

// DAGTask defines a single task in a DAG.
type DAGTask struct {
	Name         string      `json:"name"`
	Template     string      `json:"template"`
	Dependencies []string    `json:"dependencies,omitempty"`
	Arguments    *Arguments  `json:"arguments,omitempty"`
	When         string      `json:"when,omitempty"`
}

// StepGroup represents a group of parallel steps.
type StepGroup struct {
	Name      string     `json:"name"`
	Template  string     `json:"template"`
	Arguments *Arguments `json:"arguments,omitempty"`
	When      string     `json:"when,omitempty"`
}

// Arguments contains workflow or template arguments.
type Arguments struct {
	Parameters []Parameter `json:"parameters,omitempty"`
	Artifacts  []Artifact  `json:"artifacts,omitempty"`
}

// Inputs defines template inputs.
type Inputs struct {
	Parameters []Parameter `json:"parameters,omitempty"`
	Artifacts  []Artifact  `json:"artifacts,omitempty"`
}

// Outputs defines template outputs.
type Outputs struct {
	Parameters []Parameter `json:"parameters,omitempty"`
	Artifacts  []Artifact  `json:"artifacts,omitempty"`
	Result     string      `json:"result,omitempty"`
}

// Parameter defines a workflow parameter.
// Uses Go generics for type-safe values, different from Hera's Any type.
type Parameter struct {
	Name        string      `json:"name"`
	Value       interface{} `json:"value,omitempty"`
	ValueFrom   *ValueFrom  `json:"valueFrom,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Description string      `json:"description,omitempty"`
}

// ValueFrom defines a parameter value source.
type ValueFrom struct {
	Path            string `json:"path,omitempty"`
	JSONPath        string `json:"jsonPath,omitempty"`
	Expression      string `json:"expression,omitempty"`
	Parameter       string `json:"parameter,omitempty"`
}

// Artifact defines a workflow artifact.
type Artifact struct {
	Name string         `json:"name"`
	Path string         `json:"path,omitempty"`
	From string         `json:"from,omitempty"`
	S3   *S3Artifact    `json:"s3,omitempty"`
	HTTP *HTTPArtifact  `json:"http,omitempty"`
	Git  *GitArtifact   `json:"git,omitempty"`
}

// S3Artifact defines an S3 artifact location.
type S3Artifact struct {
	Endpoint string `json:"endpoint,omitempty"`
	Bucket   string `json:"bucket"`
	Key      string `json:"key"`
	Region   string `json:"region,omitempty"`
}

// HTTPArtifact defines an HTTP artifact location.
type HTTPArtifact struct {
	URL string `json:"url"`
}

// GitArtifact defines a Git artifact location.
type GitArtifact struct {
	Repo     string `json:"repo"`
	Revision string `json:"revision,omitempty"`
}

// EnvVar represents an environment variable.
type EnvVar struct {
	Name      string         `json:"name"`
	Value     string         `json:"value,omitempty"`
	ValueFrom *EnvVarSource  `json:"valueFrom,omitempty"`
}

// EnvVarSource represents a source for an environment variable value.
type EnvVarSource struct {
	SecretKeyRef   *SecretKeySelector   `json:"secretKeyRef,omitempty"`
	ConfigMapKeyRef *ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
}

// SecretKeySelector selects a key from a Secret.
type SecretKeySelector struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// ConfigMapKeySelector selects a key from a ConfigMap.
type ConfigMapKeySelector struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

// Resources defines compute resource requirements.
type Resources struct {
	Limits   ResourceList `json:"limits,omitempty"`
	Requests ResourceList `json:"requests,omitempty"`
}

// ResourceList is a map of resource name to quantity.
type ResourceList map[string]string

// WorkflowStatus represents the status of a workflow.
type WorkflowStatus struct {
	Phase      string           `json:"phase"`
	StartedAt  metav1.Time      `json:"startedAt,omitempty"`
	FinishedAt metav1.Time      `json:"finishedAt,omitempty"`
	Message    string           `json:"message,omitempty"`
	Nodes      map[string]Node  `json:"nodes,omitempty"`
}

// Node represents a workflow execution node.
type Node struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Type       string      `json:"type"`
	Phase      string      `json:"phase"`
	StartedAt  metav1.Time `json:"startedAt,omitempty"`
	FinishedAt metav1.Time `json:"finishedAt,omitempty"`
	Message    string      `json:"message,omitempty"`
}
