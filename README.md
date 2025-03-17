# hnapi

[![GoDoc](https://godoc.org/github.com/yarlson/hnapi?status.svg)](https://godoc.org/github.com/yarlson/hnapi)
[![Build Status](https://travis-ci.org/yarlson/hnapi.svg?branch=master)](https://travis-ci.org/yarlson/hnapi)

**hnapi** is a Go SDK for interacting with the [Hacker News API](https://github.com/HackerNews/API). It provides an easy-to-use interface to retrieve stories, comments, jobs, polls, and user profiles. The package also offers helper functions for batch retrieval and a real-time updates mechanism using Go channels.

## Features

- **Complete Coverage:** Fetch items (stories, comments, jobs, polls, etc.), user profiles, and lists (top, new, best, Ask, Show, Job).
- **Strongly Typed:** JSON responses are automatically parsed into Go structs.
- **Batch Retrieval:** Efficiently fetch multiple items concurrently with a configurable concurrency limit.
- **Real-Time Updates:** Subscribe to updates from the `/v0/updates` endpoint via a channel-based API.
- **Configurable & Extensible:** Customize timeouts, base URL, retry strategies, polling intervals, concurrency limits, and even inject a custom `http.Client`.
- **Context-Aware:** All methods accept `context.Context` for cancellation and deadlines.

## Installation

To install **hnapi**, run:

```bash
go get github.com/yarlson/hnapi
```

Then, import it in your project:

```go
import "github.com/yarlson/hnapi"
```

## Quick Start

Below is a simple example to get you started:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yarlson/hnapi"
)

func main() {
	// Initialize the client with custom options if needed
	client := hnapi.NewClient(
		hnapi.WithRequestTimeout(15*time.Second),
		hnapi.WithConcurrency(5),
		hnapi.WithPollInterval(30*time.Second), // Default is 30 seconds
	)

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Fetch top stories
	topStories, err := client.GetTopStories(ctx)
	if err != nil {
		log.Fatalf("Error fetching top stories: %v", err)
	}
	fmt.Printf("Top Stories: %v\n", topStories)

	// Fetch an individual item
	if len(topStories) > 0 {
		item, err := client.GetItem(ctx, topStories[0])
		if err != nil {
			log.Fatalf("Error fetching item: %v", err)
		}
		fmt.Printf("Item Details:\n Title: %s\n By: %s\n Score: %d\n", item.Title, item.By, item.Score)
	}

	// Start real-time updates
	updatesCh, err := client.StartUpdates(ctx)
	if err != nil {
		log.Fatalf("Error starting updates: %v", err)
	}

	// Process updates for a short time
	go func() {
		for update := range updatesCh {
			fmt.Printf("Updates: %d items, %d profiles changed\n", len(update.Items), len(update.Profiles))
		}
	}()

	// Wait to observe some updates (or press Ctrl+C to cancel)
	time.Sleep(1 * time.Minute)
}
```

## Configuration Options

**hnapi** uses the "with options" pattern. You can customize the client by providing various options:

- **WithBaseURL(url string):** Set a custom base URL. (Default: `https://hacker-news.firebaseio.com/v0/`)
- **WithRequestTimeout(timeout time.Duration):** Set the request timeout. (Default: 10 seconds)
- **WithMaxRetries(retries int):** Set the maximum number of retries for failed requests. (Default: 3)
- **WithBackoffInterval(interval time.Duration):** Set the backoff interval between retries. (Default: 2 seconds)
- **WithPollInterval(interval time.Duration):** Set the polling interval for real-time updates. (Default: 30 seconds)
- **WithConcurrency(concurrency int):** Set the concurrency limit for batch retrieval. (Default: 10)
- **WithHTTPClient(client \*http.Client):** Inject a custom HTTP client for advanced use cases.

Example:

```go
client := hnapi.NewClient(
	hnapi.WithBaseURL("https://custom-api.example.com/"),
	hnapi.WithRequestTimeout(15*time.Second),
	hnapi.WithConcurrency(5),
)
```

## Testing

The **hnapi** package is fully tested with unit and integration tests. To run tests, simply execute:

```bash
go test -v ./...
```

For integration tests that make real API calls, consider running them in an environment where such calls are allowed, or skip them with `-short`.

## Documentation

Full documentation is available via [GoDoc](https://godoc.org/github.com/yarlson/hnapi). Inline comments and examples are provided to ensure a smooth development experience.

## Contributing

Contributions are welcome! Please open issues or submit pull requests on [GitHub](https://github.com/yarlson/hnapi).

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
