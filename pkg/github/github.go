// Package github provides utilities and client initialization for interacting
// with the GitHub API within the MCP server.
package github

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client with additional configuration.
type Client struct {
	*github.Client
	token string
}

// ClientOption is a functional option for configuring a Client.
type ClientOption func(*Client)

// WithToken sets the authentication token for the GitHub client.
func WithToken(token string) ClientOption {
	return func(c *Client) {
		c.token = token
	}
}

// NewClient creates a new GitHub API client with the provided options.
// If no token is provided via options, it falls back to the GITHUB_TOKEN
// environment variable.
func NewClient(opts ...ClientOption) (*Client, error) {
	c := &Client{}

	for _, opt := range opts {
		opt(c)
	}

	// Fall back to environment variable if no token was set via options.
	if c.token == "" {
		c.token = os.Getenv("GITHUB_TOKEN")
	}

	if c.token == "" {
		return nil, fmt.Errorf("GitHub token is required: set GITHUB_TOKEN environment variable or provide a token")
	}

	httpClient := oauth2TokenClient(context.Background(), c.token)
	c.Client = github.NewClient(httpClient)

	return c, nil
}

// NewClientFromEnv creates a GitHub client using credentials from the environment.
// This is a convenience wrapper around NewClient.
func NewClientFromEnv() (*Client, error) {
	return NewClient()
}

// oauth2TokenClient returns an HTTP client that authenticates requests using
// the provided OAuth2 token.
func oauth2TokenClient(ctx context.Context, token string) *http.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	return oauth2.NewClient(ctx, ts)
}

// IsNotFound returns true if the error is a GitHub 404 Not Found error.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	ghErr, ok := err.(*github.ErrorResponse)
	return ok && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound
}

// IsRateLimited returns true if the error is a GitHub rate limit error.
func IsRateLimited(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*github.RateLimitError)
	return ok
}
