package hnapi_test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/yarlson/hnapi"
)

func TestIntegration(t *testing.T) {
	// Skip integration test in CI environments or when running short tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Initialize the client with custom options
	client := hnapi.NewClient(
		hnapi.WithRequestTimeout(15*time.Second),
		hnapi.WithConcurrency(5),
		hnapi.WithPollInterval(10*time.Second),
	)

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Step 1: Fetch a single item (a story)
	t.Run("GetItem", func(t *testing.T) {
		item, err := client.GetItem(ctx, 8863)
		if err != nil {
			t.Fatalf("Failed to get item: %v", err)
		}
		if item.ID != 8863 {
			t.Errorf("Expected item ID 8863, got %d", item.ID)
		}
		t.Logf("Successfully retrieved item: ID=%d, Title=%s", item.ID, item.Title)
	})

	// Step 2: Fetch a user profile
	t.Run("GetUser", func(t *testing.T) {
		user, err := client.GetUser(ctx, "pg")
		if err != nil {
			t.Fatalf("Failed to get user: %v", err)
		}
		if user.ID != "pg" {
			t.Errorf("Expected user ID 'pg', got %s", user.ID)
		}
		t.Logf("Successfully retrieved user: ID=%s, Karma=%d", user.ID, user.Karma)
	})

	// Step 3: Retrieve a list of top stories
	t.Run("GetTopStories", func(t *testing.T) {
		stories, err := client.GetTopStories(ctx)
		if err != nil {
			t.Fatalf("Failed to get top stories: %v", err)
		}
		if len(stories) == 0 {
			t.Error("Expected at least one top story, got none")
		}
		t.Logf("Successfully retrieved %d top stories", len(stories))
	})

	// Step 4: Perform batch retrieval of several items
	t.Run("GetItemsBatch", func(t *testing.T) {
		// First get some IDs to use for the batch
		stories, err := client.GetTopStories(ctx)
		if err != nil {
			t.Fatalf("Failed to get top stories: %v", err)
		}

		// Take the first 5 stories (or fewer if less than 5 are available)
		batchSize := 5
		if len(stories) < batchSize {
			batchSize = len(stories)
		}
		batchIDs := stories[:batchSize]

		// Get the batch of items
		items, err := client.GetItemsBatch(ctx, batchIDs)
		if err != nil {
			t.Fatalf("Failed to get items batch: %v", err)
		}
		if len(items) == 0 {
			t.Error("Expected at least one item in batch, got none")
		}
		t.Logf("Successfully retrieved %d items in batch", len(items))

		// Print basic info about each item
		for i, item := range items {
			t.Logf("Item %d: ID=%d, Type=%s, Title=%s", i+1, item.ID, item.Type, item.Title)
		}
	})

	// Step 5: Start real-time updates, process a few updates, and gracefully shut down
	t.Run("StartUpdates", func(t *testing.T) {
		// Create a context with a short timeout for updates
		updateCtx, updateCancel := context.WithTimeout(ctx, 15*time.Second)
		defer updateCancel()

		// Start updates
		updatesCh, err := client.StartUpdates(updateCtx)
		if err != nil {
			t.Fatalf("Failed to start updates: %v", err)
		}

		// Process updates for a short time
		updateCount := 0
		updateDeadline := time.After(10 * time.Second)

		for {
			select {
			case update, ok := <-updatesCh:
				if !ok {
					// Channel is closed, updates have stopped
					t.Logf("Updates channel closed after receiving %d updates", updateCount)
					return
				}
				updateCount++
				t.Logf("Update %d: %d items, %d profiles changed",
					updateCount, len(update.Items), len(update.Profiles))

				// Optionally, fetch a few of the updated items
				if len(update.Items) > 0 {
					itemID := update.Items[0]
					item, err := client.GetItem(ctx, itemID)
					if err != nil {
						t.Logf("Failed to get updated item %d: %v", itemID, err)
					} else {
						t.Logf("Updated item: ID=%d, Type=%s, Title=%s", item.ID, item.Type, item.Title)
					}
				}

			case <-updateDeadline:
				// We've processed updates for long enough
				t.Logf("Stopping updates after receiving %d updates", updateCount)
				updateCancel() // Cancel the context to stop polling
				return
			}
		}
	})
}

// ExampleClient demonstrates the basic usage of the Hacker News API client.
// This example doesn't run automatically because it makes real API calls.
func Example_client() {
	// Initialize client with default options
	client := hnapi.NewClient()

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get a top story
	topStories, err := client.GetTopStories(ctx)
	if err != nil {
		log.Fatalf("Failed to get top stories: %v", err)
	}

	if len(topStories) == 0 {
		log.Println("No top stories found")
		return
	}

	// Get the first story
	story, err := client.GetItem(ctx, topStories[0])
	if err != nil {
		log.Fatalf("Failed to get story: %v", err)
	}

	// Print the story details
	fmt.Printf("Top Story: %s by %s\n", story.Title, story.By)
	fmt.Printf("URL: %s\n", story.URL)
	fmt.Printf("Score: %d, Comments: %d\n", story.Score, story.Descendants)

	// Get the author's profile
	author, err := client.GetUser(ctx, story.By)
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}

	fmt.Printf("Author: %s, Karma: %d\n", author.ID, author.Karma)

	// Output:
	// (Output is omitted because it depends on live data from the Hacker News API)
}
