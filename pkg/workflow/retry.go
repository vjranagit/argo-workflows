package workflow

import (
	"time"
)

// RetryStrategy defines how to retry failed steps.
// Unlike Hera's simple retry count, this provides full backoff configuration.
type RetryStrategy struct {
	Limit              *int32        `json:"limit,omitempty"`
	RetryPolicy        string        `json:"retryPolicy,omitempty"` // Always, OnFailure, OnError
	Backoff            *Backoff      `json:"backoff,omitempty"`
	Expression         string        `json:"expression,omitempty"`
}

// Backoff defines backoff parameters for retries.
type Backoff struct {
	Duration    string  `json:"duration,omitempty"`     // e.g., "1m"
	Factor      *int32  `json:"factor,omitempty"`       // Multiplier for each retry
	MaxDuration string  `json:"maxDuration,omitempty"`  // Maximum backoff duration
}

// TimeoutPolicy defines timeout behavior for templates.
type TimeoutPolicy struct {
	Duration string `json:"duration,omitempty"` // e.g., "1h", "30m"
}

// WithRetryStrategy adds retry configuration to a template.
func WithRetryStrategy(limit int32, policy string) TemplateOption {
	return func(t *Template) {
		t.RetryStrategy = &RetryStrategy{
			Limit:       &limit,
			RetryPolicy: policy,
		}
	}
}

// WithRetryBackoff adds exponential backoff to retry strategy.
func WithRetryBackoff(duration string, factor int32, maxDuration string) TemplateOption {
	return func(t *Template) {
		if t.RetryStrategy == nil {
			limit := int32(3)
			t.RetryStrategy = &RetryStrategy{
				Limit:       &limit,
				RetryPolicy: "OnFailure",
			}
		}
		t.RetryStrategy.Backoff = &Backoff{
			Duration:    duration,
			Factor:      &factor,
			MaxDuration: maxDuration,
		}
	}
}

// WithTimeout adds timeout configuration to a template.
func WithTimeout(duration string) TemplateOption {
	return func(t *Template) {
		t.Timeout = &TimeoutPolicy{
			Duration: duration,
		}
	}
}

// RetryPolicies defines standard retry policies.
const (
	RetryPolicyAlways    = "Always"
	RetryPolicyOnFailure = "OnFailure"
	RetryPolicyOnError   = "OnError"
)

// StandardRetryStrategy creates a common retry configuration.
func StandardRetryStrategy(maxAttempts int32) *RetryStrategy {
	factor := int32(2)
	return &RetryStrategy{
		Limit:       &maxAttempts,
		RetryPolicy: RetryPolicyOnFailure,
		Backoff: &Backoff{
			Duration:    "10s",
			Factor:      &factor,
			MaxDuration: "5m",
		},
	}
}

// TimeoutDuration is a helper to create standard timeout strings.
type TimeoutDuration time.Duration

func (d TimeoutDuration) String() string {
	duration := time.Duration(d)
	if duration < time.Minute {
		return duration.String()
	}
	minutes := int(duration.Minutes())
	if minutes < 60 {
		return (time.Duration(minutes) * time.Minute).String()
	}
	hours := int(duration.Hours())
	return (time.Duration(hours) * time.Hour).String()
}

// Common timeout durations
const (
	Timeout30Seconds = TimeoutDuration(30 * time.Second)
	Timeout5Minutes  = TimeoutDuration(5 * time.Minute)
	Timeout30Minutes = TimeoutDuration(30 * time.Minute)
	Timeout1Hour     = TimeoutDuration(1 * time.Hour)
	Timeout6Hours    = TimeoutDuration(6 * time.Hour)
)
