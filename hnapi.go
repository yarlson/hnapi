// Package hnapi provides a Go SDK for interacting with the Hacker News API.
//
// This package offers a complete interface to the Hacker News API (powered by Firebase)
// including support for stories, comments, jobs, Ask HNs, Show HNs, polls, and user profiles.
// It also includes a real-time update mechanism that streams changes using Go channels.
//
// The package supports retrieving items (stories, comments, etc.), user profiles,
// various lists (top stories, new stories, etc.), and provides helper functions for
// batch retrieval and real-time updates.
package hnapi

// Version represents the current version of the hnapi package.
const Version = "0.1.0"

// Client is the Hacker News API client.
type Client struct {
	// Config contains the client configuration
	Config *Config
}

// NewClient creates a new Hacker News API client with the provided options.
func NewClient(opts ...Option) *Client {
	config := DefaultConfig()

	// Apply all provided options
	for _, opt := range opts {
		opt(config)
	}

	return &Client{
		Config: config,
	}
}

// HelloHackerNews returns a simple greeting message.
// This function is primarily used for initial testing.
func HelloHackerNews() string {
	return "Hello, Hacker News API"
}
