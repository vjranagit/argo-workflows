package client

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

// Authenticator handles authentication for Argo Workflows API.
// Different from Hera's approach - this uses interfaces for flexibility.
type Authenticator interface {
	Authenticate(req *http.Request) error
}

// BearerTokenAuth implements token-based authentication.
type BearerTokenAuth struct {
	Token string
}

// NewBearerTokenAuth creates a new bearer token authenticator.
func NewBearerTokenAuth(token string) *BearerTokenAuth {
	return &BearerTokenAuth{Token: token}
}

// Authenticate adds the bearer token to the request.
func (a *BearerTokenAuth) Authenticate(req *http.Request) error {
	if a.Token == "" {
		return fmt.Errorf("token is empty")
	}
	req.Header.Set("Authorization", "Bearer "+a.Token)
	return nil
}

// ServiceAccountAuth uses a Kubernetes service account token.
type ServiceAccountAuth struct {
	TokenPath string
	token     string
}

// NewServiceAccountAuth creates a service account authenticator.
func NewServiceAccountAuth(tokenPath string) *ServiceAccountAuth {
	if tokenPath == "" {
		tokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	}
	return &ServiceAccountAuth{TokenPath: tokenPath}
}

// Authenticate reads the service account token and adds it to the request.
func (a *ServiceAccountAuth) Authenticate(req *http.Request) error {
	if a.token == "" {
		data, err := os.ReadFile(a.TokenPath)
		if err != nil {
			return fmt.Errorf("read service account token: %w", err)
		}
		a.token = string(data)
	}

	req.Header.Set("Authorization", "Bearer "+a.token)
	return nil
}

// ArgoCLIAuth uses the Argo CLI to get a token.
// Similar to Hera's ArgoCLITokenGenerator but implemented in Go.
type ArgoCLIAuth struct {
	token string
}

// NewArgoCLIAuth creates an Argo CLI authenticator.
func NewArgoCLIAuth() *ArgoCLIAuth {
	return &ArgoCLIAuth{}
}

// Authenticate gets a token from the Argo CLI.
func (a *ArgoCLIAuth) Authenticate(req *http.Request) error {
	if a.token == "" {
		token, err := a.getTokenFromCLI()
		if err != nil {
			return fmt.Errorf("get token from argo CLI: %w", err)
		}
		a.token = token
	}

	req.Header.Set("Authorization", "Bearer "+a.token)
	return nil
}

func (a *ArgoCLIAuth) getTokenFromCLI() (string, error) {
	cmd := exec.Command("argo", "auth", "token")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("execute argo auth token: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// NoAuth is a no-op authenticator for unsecured endpoints.
type NoAuth struct{}

// NewNoAuth creates a no-op authenticator.
func NewNoAuth() *NoAuth {
	return &NoAuth{}
}

// Authenticate does nothing.
func (a *NoAuth) Authenticate(req *http.Request) error {
	return nil
}
