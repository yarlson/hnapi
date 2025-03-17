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

import (
	"net/http"
	"time"
)

// Version represents the current version of the hnapi package.
const Version = "0.1.0"

// Config defines all configurable options for the hnapi SDK.
type Config struct {
	// BaseURL is the base URL for the Hacker News API.
	BaseURL string

	// RequestTimeout is the timeout for HTTP requests.
	RequestTimeout time.Duration

	// MaxRetries is the maximum number of retries for failed requests.
	MaxRetries int

	// BackoffInterval is the time to wait between retries.
	BackoffInterval time.Duration

	// PollInterval is the time to wait between polling the updates endpoint.
	PollInterval time.Duration

	// Concurrency is the maximum number of concurrent requests for batch operations.
	Concurrency int

	// HTTPClient is the HTTP client used for making requests.
	HTTPClient *http.Client
}

// DefaultConfig returns a default configuration for the Hacker News API client.
func DefaultConfig() *Config {
	return &Config{
		BaseURL:         "https://hacker-news.firebaseio.com/v0/",
		RequestTimeout:  10 * time.Second,
		MaxRetries:      3,
		BackoffInterval: 2 * time.Second,
		PollInterval:    30 * time.Second,
		Concurrency:     10,
		HTTPClient:      http.DefaultClient,
	}
}

// Option is a function that modifies the Config.
type Option func(*Config)

// WithBaseURL sets a custom base URL.
func WithBaseURL(url string) Option {
	return func(c *Config) {
		c.BaseURL = url
	}
}

// WithRequestTimeout sets a custom request timeout.
func WithRequestTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.RequestTimeout = timeout
	}
}

// WithMaxRetries sets a custom maximum number of retries.
func WithMaxRetries(retries int) Option {
	return func(c *Config) {
		c.MaxRetries = retries
	}
}

// WithBackoffInterval sets a custom backoff interval between retries.
func WithBackoffInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.BackoffInterval = interval
	}
}

// WithPollInterval sets a custom polling interval for updates.
func WithPollInterval(interval time.Duration) Option {
	return func(c *Config) {
		c.PollInterval = interval
	}
}

// WithConcurrency sets a custom concurrency limit for batch operations.
func WithConcurrency(concurrency int) Option {
	return func(c *Config) {
		c.Concurrency = concurrency
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) {
		c.HTTPClient = client
	}
}

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
