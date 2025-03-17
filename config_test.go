package hnapi

import (
	"net/http"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.BaseURL != "https://hacker-news.firebaseio.com/v0/" {
		t.Errorf("Expected BaseURL to be %q, got %q", "https://hacker-news.firebaseio.com/v0/", config.BaseURL)
	}

	if config.RequestTimeout != 10*time.Second {
		t.Errorf("Expected RequestTimeout to be %v, got %v", 10*time.Second, config.RequestTimeout)
	}

	if config.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries to be %d, got %d", 3, config.MaxRetries)
	}

	if config.BackoffInterval != 2*time.Second {
		t.Errorf("Expected BackoffInterval to be %v, got %v", 2*time.Second, config.BackoffInterval)
	}

	if config.PollInterval != 30*time.Second {
		t.Errorf("Expected PollInterval to be %v, got %v", 30*time.Second, config.PollInterval)
	}

	if config.Concurrency != 10 {
		t.Errorf("Expected Concurrency to be %d, got %d", 10, config.Concurrency)
	}

	if config.HTTPClient != http.DefaultClient {
		t.Errorf("Expected HTTPClient to be http.DefaultClient")
	}
}

func TestConfigWithOptions(t *testing.T) {
	customURL := "https://custom-api.example.com/"
	customTimeout := 5 * time.Second
	customRetries := 5
	customBackoff := 1 * time.Second
	customPollInterval := 15 * time.Second
	customConcurrency := 20
	customClient := &http.Client{Timeout: 30 * time.Second}

	client := NewClient(
		WithBaseURL(customURL),
		WithRequestTimeout(customTimeout),
		WithMaxRetries(customRetries),
		WithBackoffInterval(customBackoff),
		WithPollInterval(customPollInterval),
		WithConcurrency(customConcurrency),
		WithHTTPClient(customClient),
	)

	config := client.Config

	// Verify that all options were correctly applied
	if config.BaseURL != customURL {
		t.Errorf("Expected BaseURL to be %q, got %q", customURL, config.BaseURL)
	}

	if config.RequestTimeout != customTimeout {
		t.Errorf("Expected RequestTimeout to be %v, got %v", customTimeout, config.RequestTimeout)
	}

	if config.MaxRetries != customRetries {
		t.Errorf("Expected MaxRetries to be %d, got %d", customRetries, config.MaxRetries)
	}

	if config.BackoffInterval != customBackoff {
		t.Errorf("Expected BackoffInterval to be %v, got %v", customBackoff, config.BackoffInterval)
	}

	if config.PollInterval != customPollInterval {
		t.Errorf("Expected PollInterval to be %v, got %v", customPollInterval, config.PollInterval)
	}

	if config.Concurrency != customConcurrency {
		t.Errorf("Expected Concurrency to be %d, got %d", customConcurrency, config.Concurrency)
	}

	if config.HTTPClient != customClient {
		t.Errorf("Expected HTTPClient to be the custom client")
	}
}

func TestPartialOptions(t *testing.T) {
	// Test with only some options provided
	customTimeout := 7 * time.Second
	customConcurrency := 15

	client := NewClient(
		WithRequestTimeout(customTimeout),
		WithConcurrency(customConcurrency),
	)

	config := client.Config

	// Verify that specified options were applied
	if config.RequestTimeout != customTimeout {
		t.Errorf("Expected RequestTimeout to be %v, got %v", customTimeout, config.RequestTimeout)
	}

	if config.Concurrency != customConcurrency {
		t.Errorf("Expected Concurrency to be %d, got %d", customConcurrency, config.Concurrency)
	}

	// Verify that unspecified options retain their default values
	defaultConfig := DefaultConfig()

	if config.BaseURL != defaultConfig.BaseURL {
		t.Errorf("Expected BaseURL to be default %q, got %q", defaultConfig.BaseURL, config.BaseURL)
	}

	if config.MaxRetries != defaultConfig.MaxRetries {
		t.Errorf("Expected MaxRetries to be default %d, got %d", defaultConfig.MaxRetries, config.MaxRetries)
	}

	if config.BackoffInterval != defaultConfig.BackoffInterval {
		t.Errorf("Expected BackoffInterval to be default %v, got %v", defaultConfig.BackoffInterval, config.BackoffInterval)
	}

	if config.PollInterval != defaultConfig.PollInterval {
		t.Errorf("Expected PollInterval to be default %v, got %v", defaultConfig.PollInterval, config.PollInterval)
	}
}
