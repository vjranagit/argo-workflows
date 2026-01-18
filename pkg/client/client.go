package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vjranagit/argo-workflows/pkg/workflow"
)

// Client defines the interface for interacting with Argo Workflows.
// Unlike Hera's requests-based client, this uses native Go net/http
// with proper context support for cancellation.
type Client interface {
	CreateWorkflow(ctx context.Context, wf *workflow.Workflow) (*workflow.WorkflowStatus, error)
	GetWorkflow(ctx context.Context, namespace, name string) (*workflow.Workflow, error)
	ListWorkflows(ctx context.Context, namespace string, opts ListOptions) (*WorkflowList, error)
	DeleteWorkflow(ctx context.Context, namespace, name string) error
	WatchWorkflow(ctx context.Context, namespace, name string) (<-chan WorkflowEvent, error)
}

// HTTPClient implements Client using HTTP/REST API.
type HTTPClient struct {
	baseURL    string
	namespace  string
	auth       Authenticator
	httpClient *http.Client
}

// Config holds configuration for the HTTP client.
type Config struct {
	BaseURL    string
	Namespace  string
	Auth       Authenticator
	Timeout    time.Duration
	Insecure   bool
}

// NewHTTPClient creates a new HTTP client for Argo Workflows.
func NewHTTPClient(cfg Config) *HTTPClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &HTTPClient{
		baseURL:   strings.TrimSuffix(cfg.BaseURL, "/"),
		namespace: cfg.Namespace,
		auth:      cfg.Auth,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// CreateWorkflow submits a new workflow to Argo.
func (c *HTTPClient) CreateWorkflow(ctx context.Context, wf *workflow.Workflow) (*workflow.WorkflowStatus, error) {
	if wf.Namespace == "" {
		wf.Namespace = c.namespace
	}

	// Set TypeMeta
	wf.APIVersion = "argoproj.io/v1alpha1"
	wf.Kind = "Workflow"

	body, err := json.Marshal(wf)
	if err != nil {
		return nil, fmt.Errorf("marshal workflow: %w", err)
	}

	url := fmt.Sprintf("%s/api/v1/workflows/%s", c.baseURL, wf.Namespace)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if c.auth != nil {
		if err := c.auth.Authenticate(req); err != nil {
			return nil, fmt.Errorf("authenticate: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var result workflow.Workflow
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result.Status, nil
}

// GetWorkflow retrieves a workflow by name.
func (c *HTTPClient) GetWorkflow(ctx context.Context, namespace, name string) (*workflow.Workflow, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	url := fmt.Sprintf("%s/api/v1/workflows/%s/%s", c.baseURL, namespace, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if c.auth != nil {
		if err := c.auth.Authenticate(req); err != nil {
			return nil, fmt.Errorf("authenticate: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var wf workflow.Workflow
	if err := json.NewDecoder(resp.Body).Decode(&wf); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &wf, nil
}

// ListOptions contains options for listing workflows.
type ListOptions struct {
	LabelSelector string
	FieldSelector string
	Limit         int64
	Continue      string
}

// WorkflowList represents a list of workflows.
type WorkflowList struct {
	Items    []workflow.Workflow `json:"items"`
	Metadata ListMetadata         `json:"metadata"`
}

// ListMetadata contains metadata about a list response.
type ListMetadata struct {
	Continue        string `json:"continue,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`
}

// ListWorkflows lists workflows in a namespace.
func (c *HTTPClient) ListWorkflows(ctx context.Context, namespace string, opts ListOptions) (*WorkflowList, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	url := fmt.Sprintf("%s/api/v1/workflows/%s", c.baseURL, namespace)

	// Add query parameters
	if opts.LabelSelector != "" || opts.FieldSelector != "" || opts.Limit > 0 {
		params := make([]string, 0)
		if opts.LabelSelector != "" {
			params = append(params, "labelSelector="+opts.LabelSelector)
		}
		if opts.FieldSelector != "" {
			params = append(params, "fieldSelector="+opts.FieldSelector)
		}
		if opts.Limit > 0 {
			params = append(params, fmt.Sprintf("limit=%d", opts.Limit))
		}
		if opts.Continue != "" {
			params = append(params, "continue="+opts.Continue)
		}
		if len(params) > 0 {
			url += "?" + strings.Join(params, "&")
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if c.auth != nil {
		if err := c.auth.Authenticate(req); err != nil {
			return nil, fmt.Errorf("authenticate: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var list WorkflowList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &list, nil
}

// DeleteWorkflow deletes a workflow.
func (c *HTTPClient) DeleteWorkflow(ctx context.Context, namespace, name string) error {
	if namespace == "" {
		namespace = c.namespace
	}

	url := fmt.Sprintf("%s/api/v1/workflows/%s/%s", c.baseURL, namespace, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	if c.auth != nil {
		if err := c.auth.Authenticate(req); err != nil {
			return fmt.Errorf("authenticate: %w", err)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// WorkflowEvent represents a workflow watch event.
type WorkflowEvent struct {
	Type     string             `json:"type"`
	Workflow *workflow.Workflow `json:"object"`
}

// WatchWorkflow watches for workflow events.
// Uses Go channels for event streaming, different from Hera's approach.
func (c *HTTPClient) WatchWorkflow(ctx context.Context, namespace, name string) (<-chan WorkflowEvent, error) {
	if namespace == "" {
		namespace = c.namespace
	}

	events := make(chan WorkflowEvent)

	go func() {
		defer close(events)

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		var lastPhase string

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				wf, err := c.GetWorkflow(ctx, namespace, name)
				if err != nil {
					continue
				}

				if wf.Status.Phase != lastPhase {
					events <- WorkflowEvent{
						Type:     "MODIFIED",
						Workflow: wf,
					}
					lastPhase = wf.Status.Phase
				}

				// Stop watching if workflow is complete
				if wf.Status.Phase == "Succeeded" || wf.Status.Phase == "Failed" || wf.Status.Phase == "Error" {
					return
				}
			}
		}
	}()

	return events, nil
}
