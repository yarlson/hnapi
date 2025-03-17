package hnapi

import (
	"net/http"
	"time"
)

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
