package hnapi

import (
	"context"
	"fmt"
	"log"
	"time"
)

// StartUpdates begins polling the updates endpoint and returns a channel of Updates.
// It uses the client's PollInterval configuration to determine the polling frequency.
// The polling will continue until the provided context is canceled.
//
// The returned channel will be closed when the context is canceled or if an unrecoverable
// error occurs.
func (c *Client) StartUpdates(ctx context.Context) (<-chan Updates, error) {
	// Create a buffered channel to send updates through
	// We use a buffer of 1 to ensure that a slow consumer doesn't block the polling
	updatesCh := make(chan Updates, 1)

	// Start a goroutine for polling
	go func() {
		defer close(updatesCh)

		// Create a ticker with the configured poll interval
		ticker := time.NewTicker(c.Config.PollInterval)
		defer ticker.Stop()

		// Poll immediately on start, then wait for ticker
		if err := c.pollUpdates(ctx, updatesCh); err != nil {
			// Log the error but continue polling
			log.Printf("Error polling updates: %v", err)
		}

		// Main polling loop
		for {
			select {
			case <-ctx.Done():
				// Context was canceled, stop polling
				return
			case <-ticker.C:
				// Time to poll again
				if err := c.pollUpdates(ctx, updatesCh); err != nil {
					// Log the error but continue polling
					log.Printf("Error polling updates: %v", err)
				}
			}
		}
	}()

	return updatesCh, nil
}

// pollUpdates fetches the latest updates from the API and sends them to the updates channel.
func (c *Client) pollUpdates(ctx context.Context, updatesCh chan<- Updates) error {
	// Fetch updates from the API
	var updates Updates
	if err := c.makeRequest(ctx, "updates.json", &updates); err != nil {
		return fmt.Errorf("failed to get updates: %w", err)
	}

	// Only send updates if there are any
	if len(updates.Items) > 0 || len(updates.Profiles) > 0 {
		// Try to send updates, but respect context cancellation
		select {
		case updatesCh <- updates:
			// Successfully sent updates
		case <-ctx.Done():
			// Context was canceled
			return ctx.Err()
		}
	}

	return nil
}
