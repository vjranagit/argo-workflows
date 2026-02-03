package workflow

import (
	"testing"
)

func TestRetryStrategy(t *testing.T) {
	tests := []struct {
		name     string
		limit    int32
		policy   string
		wantNil  bool
	}{
		{
			name:    "standard retry on failure",
			limit:   3,
			policy:  RetryPolicyOnFailure,
			wantNil: false,
		},
		{
			name:    "always retry",
			limit:   5,
			policy:  RetryPolicyAlways,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := &Template{Name: "test"}
			opt := WithRetryStrategy(tt.limit, tt.policy)
			opt(tmpl)

			if (tmpl.RetryStrategy == nil) != tt.wantNil {
				t.Errorf("RetryStrategy nil = %v, want %v", tmpl.RetryStrategy == nil, tt.wantNil)
			}

			if tmpl.RetryStrategy != nil {
				if *tmpl.RetryStrategy.Limit != tt.limit {
					t.Errorf("Limit = %v, want %v", *tmpl.RetryStrategy.Limit, tt.limit)
				}
				if tmpl.RetryStrategy.RetryPolicy != tt.policy {
					t.Errorf("Policy = %v, want %v", tmpl.RetryStrategy.RetryPolicy, tt.policy)
				}
			}
		})
	}
}

func TestRetryBackoff(t *testing.T) {
	tmpl := &Template{Name: "test"}
	
	// Apply backoff option
	opt := WithRetryBackoff("10s", 2, "5m")
	opt(tmpl)

	if tmpl.RetryStrategy == nil {
		t.Fatal("RetryStrategy should be initialized")
	}

	if tmpl.RetryStrategy.Backoff == nil {
		t.Fatal("Backoff should be set")
	}

	if tmpl.RetryStrategy.Backoff.Duration != "10s" {
		t.Errorf("Duration = %v, want 10s", tmpl.RetryStrategy.Backoff.Duration)
	}

	if *tmpl.RetryStrategy.Backoff.Factor != 2 {
		t.Errorf("Factor = %v, want 2", *tmpl.RetryStrategy.Backoff.Factor)
	}

	if tmpl.RetryStrategy.Backoff.MaxDuration != "5m" {
		t.Errorf("MaxDuration = %v, want 5m", tmpl.RetryStrategy.Backoff.MaxDuration)
	}
}

func TestTimeout(t *testing.T) {
	tmpl := &Template{Name: "test"}
	
	opt := WithTimeout("30m")
	opt(tmpl)

	if tmpl.Timeout == nil {
		t.Fatal("Timeout should be set")
	}

	if tmpl.Timeout.Duration != "30m" {
		t.Errorf("Duration = %v, want 30m", tmpl.Timeout.Duration)
	}
}

func TestStandardRetryStrategy(t *testing.T) {
	strategy := StandardRetryStrategy(5)

	if strategy == nil {
		t.Fatal("strategy should not be nil")
	}

	if *strategy.Limit != 5 {
		t.Errorf("Limit = %v, want 5", *strategy.Limit)
	}

	if strategy.RetryPolicy != RetryPolicyOnFailure {
		t.Errorf("Policy = %v, want %v", strategy.RetryPolicy, RetryPolicyOnFailure)
	}

	if strategy.Backoff == nil {
		t.Fatal("Backoff should be initialized")
	}

	if strategy.Backoff.Duration != "10s" {
		t.Errorf("Duration = %v, want 10s", strategy.Backoff.Duration)
	}
}

func TestTimeoutDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration TimeoutDuration
		want     string
	}{
		{"30 seconds", Timeout30Seconds, "30s"},
		{"5 minutes", Timeout5Minutes, "5m0s"},
		{"1 hour", Timeout1Hour, "1h0m0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.duration.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainerTemplateWithRetry(t *testing.T) {
	// Test that retry and timeout options work with ContainerTemplate
	tmpl := ContainerTemplate(
		"retry-container",
		WithImage("alpine:latest"),
		WithCommand("sh", "-c", "exit 1"),
		WithRetryStrategy(3, RetryPolicyOnFailure),
		WithTimeout("5m"),
	)

	if tmpl.Name != "retry-container" {
		t.Errorf("Name = %v, want retry-container", tmpl.Name)
	}

	if tmpl.RetryStrategy == nil {
		t.Fatal("RetryStrategy should be set")
	}

	if *tmpl.RetryStrategy.Limit != 3 {
		t.Errorf("Limit = %v, want 3", *tmpl.RetryStrategy.Limit)
	}

	if tmpl.Timeout == nil {
		t.Fatal("Timeout should be set")
	}

	if tmpl.Timeout.Duration != "5m" {
		t.Errorf("Duration = %v, want 5m", tmpl.Timeout.Duration)
	}
}
