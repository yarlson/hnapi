# Hacker News API SDK for Go

A complete Go SDK for interacting with the [Hacker News API](https://github.com/HackerNews/API) (powered by Firebase).

## Overview

This SDK provides a comprehensive interface to the Hacker News API, including support for:

- Stories, comments, jobs, Ask HNs, Show HNs, and polls
- User profiles
- Real-time updates via Go channels
- Batch retrieval with concurrent requests

## Installation

```bash
go get github.com/yarlson/hnapi
```

## Usage

### Initializing the Client

```go
import "github.com/yarlson/hnapi"

// Use default configuration
client := hnapi.NewClient()

// Or customize with options
client := hnapi.NewClient(
    hnapi.WithRequestTimeout(5 * time.Second),
    hnapi.WithConcurrency(20),
)
```

### Retrieving Items

```go
// Get a story, comment, or other item by ID
item, err := client.GetItem(context.Background(), 8863)
if err != nil {
    // Handle error
}
fmt.Printf("Story: %s by %s\n", item.Title, item.By)
```

### Retrieving User Profiles

```go
user, err := client.GetUser(context.Background(), "pg")
if err != nil {
    // Handle error
}
fmt.Printf("User: %s, Karma: %d\n", user.ID, user.Karma)
```

### Getting Lists of Stories

```go
// Get top stories
topStories, err := client.GetTopStories(context.Background())
if err != nil {
    // Handle error
}

// Similar functions exist for:
// - GetNewStories()
// - GetBestStories()
// - GetAskStories()
// - GetShowStories()
// - GetJobStories()
```

### Batch Retrieval

```go
// Get multiple items concurrently
items, err := client.GetItemsBatch(context.Background(), []int{8863, 8864, 8865})
if err != nil {
    // Handle error
}
for _, item := range items {
    fmt.Printf("Item %d: %s\n", item.ID, item.Title)
}
```

### Real-time Updates

```go
// Start receiving updates
ctx, cancel := context.WithCancel(context.Background())
defer cancel() // Ensure the update stream is stopped when done

updatesCh, err := client.StartUpdates(ctx)
if err != nil {
    // Handle error
}

// Process updates as they arrive
for updates := range updatesCh {
    for _, itemID := range updates.Items {
        // Process updated items
    }
    for _, userID := range updates.Profiles {
        // Process updated user profiles
    }
}
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. 