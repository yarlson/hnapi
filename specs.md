# Hacker News Go SDK Specification

## Overview

The Hacker News Go SDK provides a complete interface to the [Hacker News API](https://github.com/HackerNews/API) (powered by Firebase) including support for stories, comments, jobs, Ask HNs, Show HNs, polls (and pollopts), and user profiles. In addition, the SDK includes a real-time update mechanism that streams changes using Go channels.

---

## 1. Core Functionalities

### 1.1 Endpoints Covered
- **Items:** Retrieve and parse items such as stories, comments, jobs, polls, pollopts, etc.
    - Endpoint: `/v0/item/<id>.json`
- **Users:** Retrieve user profiles with fields like `id`, `created`, `karma`, `about`, and `submitted`.
    - Endpoint: `/v0/user/<username>.json`
- **Lists:** Support retrieval of top stories, new stories, best stories, Ask HNs, Show HNs, and Job stories.
    - Endpoints: `/v0/topstories`, `/v0/newstories`, `/v0/beststories`, `/v0/askstories`, `/v0/showstories`, `/v0/jobstories`
- **Live Data:**
    - **Max Item ID:** Retrieve the latest item ID from `/v0/maxitem`.
    - **Updates:** Retrieve changes for items and profiles from `/v0/updates`.

### 1.2 Real-Time Updates
- **Mechanism:** The SDK will poll the `/v0/updates` endpoint periodically.
- **Channel-based API:** Updates are streamed through a Go channel.
- **Polling Interval:** Configurable via options; default is **30 seconds**.
- **Graceful Shutdown:** Use context cancellation to stop polling gracefully.

### 1.3 Batch Retrieval
- **Concurrent Fetching:** Provide helper functions to retrieve multiple items concurrently using goroutines.
- **Concurrency Limit:** Default concurrency limit is **10** (overridable via options).

---

## 2. Architecture & Design Choices

### 2.1 Package Structure
- **Single Package:** All functionalities reside in a single package (e.g., `hackernews`). This package will encapsulate:
    - API interactions.
    - Real-time update streaming.
    - Data parsing and helper functions.

### 2.2 Options Pattern
- **Configuration Struct:** The SDK exposes a configuration struct using the "with options" pattern to allow developers to set:
    - **Base URL:** The API’s base endpoint (default: `https://hacker-news.firebaseio.com/v0/`).
    - **Request Timeout:** Timeout for HTTP requests.
    - **Retry Strategy:** Including max retries and backoff intervals.
    - **Polling Interval:** For real-time updates (default: 30 seconds).
    - **Concurrency Limit:** For batch retrieval (default: 10).
    - **Custom HTTP Client:** Allow injection of a custom `http.Client` for advanced use cases (middleware, proxies, etc.).

### 2.3 Context Support
- **Context Parameters:** All public SDK methods should accept a `context.Context` to support cancellation and timeouts in an idiomatic way.

---

## 3. Data Handling

### 3.1 Strongly-Typed Go Structs
- **Items:** Parse JSON responses into strongly-typed structs for each item type.
- **User Profiles:** Parse JSON responses into a user struct.

#### Example Structs

```go
// Item represents a generic Hacker News item.
type Item struct {
    ID          int      `json:"id"`
    Deleted     bool     `json:"deleted,omitempty"`
    Type        string   `json:"type"`
    By          string   `json:"by,omitempty"`
    Time        int64    `json:"time"`
    Text        string   `json:"text,omitempty"`
    Dead        bool     `json:"dead,omitempty"`
    Parent      int      `json:"parent,omitempty"`
    Poll        int      `json:"poll,omitempty"`
    Kids        []int    `json:"kids,omitempty"`
    URL         string   `json:"url,omitempty"`
    Score       int      `json:"score,omitempty"`
    Title       string   `json:"title,omitempty"`
    Parts       []int    `json:"parts,omitempty"`
    Descendants int      `json:"descendants,omitempty"`
}

// User represents a Hacker News user.
type User struct {
    ID        string   `json:"id"`
    Created   int64    `json:"created"`
    Karma     int      `json:"karma"`
    About     string   `json:"about,omitempty"`
    Submitted []int    `json:"submitted,omitempty"`
}
```

### 3.2 On-Demand Data Retrieval
- **No Caching:** All data retrieval will be performed on-demand. There is no built-in caching mechanism.

---

## 4. Error Handling

- **Return Errors:** All SDK functions should return errors to the caller for handling.
- **Error Propagation:** When a network call or JSON parsing fails, propagate errors with context-specific messages.
- **Retries:** In case of transient errors, implement a configurable retry strategy.

---

## 5. Real-Time Updates Implementation

### 5.1 Polling Mechanism
- **Goroutine:** Launch a goroutine that polls the `/v0/updates` endpoint at the specified polling interval.
- **Channel Interface:** Stream updates (both item IDs and user profiles) through a Go channel.
- **Context Cancellation:** The polling goroutine should listen for context cancellation and shutdown gracefully.

#### Pseudocode Example

```go
func (client *Client) StartUpdates(ctx context.Context) (<-chan Updates, error) {
    updatesCh := make(chan Updates)
    go func() {
        defer close(updatesCh)
        ticker := time.NewTicker(client.config.PollInterval)
        defer ticker.Stop()
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                updates, err := client.fetchUpdates(ctx)
                if err != nil {
                    // Handle or propagate the error as needed
                    continue
                }
                updatesCh <- updates
            }
        }
    }()
    return updatesCh, nil
}
```

### 5.2 Update Struct
```go
// Updates represents the structure of the /v0/updates endpoint.
type Updates struct {
    Items    []int    `json:"items"`
    Profiles []string `json:"profiles"`
}
```

---

## 6. HTTP Client Customization

- **Custom HTTP Client:** Allow injection of a custom `http.Client` via the options pattern.
- **Default Client:** If none is provided, use the default `http.Client` with the configured timeout.

---

## 7. Testing Plan

### 7.1 Unit Tests
- **Coverage:** Write unit tests for:
    - API endpoint interactions (e.g., item retrieval, user profiles, list endpoints).
    - Parsing logic for JSON responses into Go structs.
    - Batch retrieval functionality ensuring concurrency limits.
    - Real-time updates polling logic (simulate API responses).
    - Options configuration and default values.
- **Mocking:** Use mocking of HTTP responses (e.g., via `httptest.Server`) to simulate various API behaviors and errors.

### 7.2 Example Code Snippets & Documentation
- **GoDoc:** Provide extensive inline comments for all exported functions, types, and methods.
- **README:** Include a separate README file with:
    - Detailed configuration options.
    - Step-by-step usage examples for retrieving items, batch processing, and subscribing to real-time updates.
    - Example usage of context cancellation for graceful shutdown.

---

## 8. Full API Schema

### Endpoints & Expected JSON Schemas

#### Item Endpoint
- **URI:** `https://hacker-news.firebaseio.com/v0/item/<id>.json`
- **Example Response:**
  ```json
  {
    "by": "dhouston",
    "descendants": 71,
    "id": 8863,
    "kids": [8952, 9224, ...],
    "score": 111,
    "time": 1175714200,
    "title": "My YC app: Dropbox - Throw away your USB drive",
    "type": "story",
    "url": "http://www.getdropbox.com/u/2/screencast.html"
  }
  ```

#### User Endpoint
- **URI:** `https://hacker-news.firebaseio.com/v0/user/<username>.json`
- **Example Response:**
  ```json
  {
    "about": "This is a test",
    "created": 1173923446,
    "id": "jl",
    "karma": 2937,
    "submitted": [8265435, 8168423, ...]
  }
  ```

#### List Endpoints
- **Top Stories:** `https://hacker-news.firebaseio.com/v0/topstories.json`
- **New Stories:** `https://hacker-news.firebaseio.com/v0/newstories.json`
- **Best Stories:** `https://hacker-news.firebaseio.com/v0/beststories.json`
- **Ask Stories:** `https://hacker-news.firebaseio.com/v0/askstories.json`
- **Show Stories:** `https://hacker-news.firebaseio.com/v0/showstories.json`
- **Job Stories:** `https://hacker-news.firebaseio.com/v0/jobstories.json`

#### Max Item Endpoint
- **URI:** `https://hacker-news.firebaseio.com/v0/maxitem.json`
- **Example Response:**
  ```json
  9130260
  ```

#### Updates Endpoint
- **URI:** `https://hacker-news.firebaseio.com/v0/updates.json`
- **Example Response:**
  ```json
  {
    "items": [8423305, 8420805, ...],
    "profiles": ["thefox", "mdda", ...]
  }
  ```

---

## 9. Example Configuration Struct and Client Initialization

```go
package hackernews

import (
    "context"
    "net/http"
    "time"
)

// Config defines all configurable options for the SDK.
type Config struct {
    BaseURL         string
    RequestTimeout  time.Duration
    MaxRetries      int
    BackoffInterval time.Duration
    PollInterval    time.Duration
    Concurrency     int
    HTTPClient      *http.Client
}

// DefaultConfig returns a default configuration.
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

// Option is a functional option for configuring the client.
type Option func(*Config)

// WithRequestTimeout sets a custom request timeout.
func WithRequestTimeout(timeout time.Duration) Option {
    return func(c *Config) {
        c.RequestTimeout = timeout
    }
}

// Other With* functions can be similarly defined.

type Client struct {
    config *Config
    // other fields as necessary (e.g., for caching, logging)
}

// NewClient creates a new Hacker News client with the provided options.
func NewClient(opts ...Option) *Client {
    config := DefaultConfig()
    for _, opt := range opts {
        opt(config)
    }
    return &Client{config: config}
}
```

---

## 10. Conclusion

This specification covers all aspects of the Hacker News Go SDK, including:

- **Core Endpoints:** Items, Users, Lists, Max Item, and Updates.
- **Real-Time Updates:** Streaming via Go channels with configurable polling and graceful shutdown via context cancellation.
- **Options Pattern:** Comprehensive configuration for request timeout, base URL, retry strategies, polling interval, concurrency, and custom HTTP client.
- **Data Handling:** Strongly-typed structs for JSON responses.
- **Batch Retrieval:** Concurrent fetching with a default concurrency limit.
- **Error Handling:** Return errors to the caller.
- **Testing:** A robust unit testing plan using mocks and examples.
- **Documentation:** Extensive GoDoc comments and a separate README with usage examples.

A developer can now immediately begin the implementation using this specification as a blueprint. If there are any additional questions or refinements needed during implementation, further discussions can be scheduled.
