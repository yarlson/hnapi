package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yarlson/hnapi"
)

func main() {
	fmt.Println("Hacker News API Example")
	fmt.Println("======================")
	fmt.Println("Press Ctrl+C at any time to exit")

	// Initialize client with custom options
	client := hnapi.NewClient(
		hnapi.WithRequestTimeout(15*time.Second),
		hnapi.WithConcurrency(5),
		hnapi.WithPollInterval(5*time.Second), // Short poll interval for demo
	)

	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling for graceful shutdown
	// This channel will be closed when the program should exit
	done := make(chan struct{})

	// Improved signal handling with a separate goroutine
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		fmt.Printf("\n\nReceived signal %v, terminating...\n", sig)
		cancel()    // Cancel the context
		close(done) // Signal that we should exit
	}()

	// Step 1: Fetch top stories
	fmt.Println("\n🔍 Fetching top stories...")
	topStories, err := client.GetTopStories(ctx)
	if err != nil {
		log.Fatalf("Failed to get top stories: %v", err)
	}
	fmt.Printf("Got %d top stories\n", len(topStories))

	// Step 2: Get first story details
	if len(topStories) > 0 {
		fmt.Printf("\n📰 Fetching details for top story (ID: %d)...\n", topStories[0])
		story, err := client.GetItem(ctx, topStories[0])
		if err != nil {
			log.Fatalf("Failed to get story: %v", err)
		}

		fmt.Printf("Title: %s\n", story.Title)
		fmt.Printf("By: %s\n", story.By)
		fmt.Printf("Score: %d\n", story.Score)
		fmt.Printf("URL: %s\n", story.URL)
		fmt.Printf("Comments: %d\n", story.Descendants)

		// Step 3: Fetch the author's profile
		fmt.Printf("\n👤 Fetching profile for user %s...\n", story.By)
		author, err := client.GetUser(ctx, story.By)
		if err != nil {
			log.Printf("Failed to get user: %v", err)
		} else {
			fmt.Printf("User: %s\n", author.ID)
			fmt.Printf("Karma: %d\n", author.Karma)
			fmt.Printf("Created: %s\n", time.Unix(author.Created, 0).Format(time.RFC3339))
		}
	}

	// Step 4: Batch retrieval
	batchSize := 5
	if len(topStories) < batchSize {
		batchSize = len(topStories)
	}
	batchIDs := topStories[:batchSize]

	fmt.Printf("\n📚 Fetching %d stories in batch...\n", len(batchIDs))
	items, err := client.GetItemsBatch(ctx, batchIDs)
	if err != nil {
		log.Printf("Batch retrieval had errors: %v", err)
	}
	fmt.Printf("Successfully retrieved %d/%d items\n", len(items), len(batchIDs))

	// Print a summary of each item
	for i, item := range items {
		fmt.Printf("%d. %s by %s (%d points)\n", i+1, item.Title, item.By, item.Score)
	}

	// Step 5: Real-time updates
	fmt.Println("\n🔄 Starting real-time updates (press Ctrl+C to exit)...")
	updatesCh, err := client.StartUpdates(ctx)
	if err != nil {
		log.Fatalf("Failed to start updates: %v", err)
	}

	updateCount := 0
	updateTicker := time.NewTicker(100 * time.Millisecond) // Fast ticker to ensure responsiveness
	defer updateTicker.Stop()

	// Main processing loop with improved termination handling
	running := true
	for running {
		select {
		case update, ok := <-updatesCh:
			if !ok {
				fmt.Println("Updates channel closed")
				running = false
				break
			}
			updateCount++
			fmt.Printf("Update %d: %d items, %d profiles changed\n",
				updateCount, len(update.Items), len(update.Profiles))

			// If we got item updates, fetch the first one as an example
			if len(update.Items) > 0 {
				itemID := update.Items[0]
				item, err := client.GetItem(ctx, itemID)
				if err != nil {
					fmt.Printf("Failed to get updated item %d: %v\n", itemID, err)
				} else {
					fmt.Printf("Updated item: ID=%d, Type=%s, Title=%s\n",
						item.ID, item.Type, item.Title)
				}
			}

		case <-ctx.Done():
			fmt.Printf("\nContext canceled, stopping updates\n")
			running = false
			break

		case <-done:
			// Exit signal received
			fmt.Printf("\nExiting after receiving %d updates\n", updateCount)
			running = false
			break

		case <-updateTicker.C:
			// This case ensures we check for termination regularly
			// even if no updates are coming in
			if ctx.Err() != nil {
				running = false
				break
			}
		}
	}

	fmt.Println("Example application terminated gracefully")
}
