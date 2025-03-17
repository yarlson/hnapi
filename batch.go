package hnapi

import (
	"context"
	"fmt"
	"sync"
)

// GetItemsBatch retrieves multiple items concurrently by their IDs.
// It respects the client's Concurrency configuration to limit the number of concurrent requests.
func (c *Client) GetItemsBatch(ctx context.Context, ids []int) ([]*Item, error) {
	if len(ids) == 0 {
		return []*Item{}, nil
	}

	// Create a context that we can cancel if needed
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Channel to collect results
	resultCh := make(chan *itemResult, len(ids))

	// Use a semaphore to limit concurrency
	sem := make(chan struct{}, c.Config.Concurrency)

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start a goroutine for each item ID
	for _, id := range ids {
		// Add to wait group before spawning goroutine
		wg.Add(1)

		go func(id int) {
			defer wg.Done()

			// Acquire a token from the semaphore
			sem <- struct{}{}
			defer func() { <-sem }() // Release the token when done

			// Get the item
			item, err := c.GetItem(ctx, id)

			// Send the result through the channel
			resultCh <- &itemResult{
				Item:  item,
				ID:    id,
				Error: err,
			}
		}(id)
	}

	// Close the results channel once all goroutines are done
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	items := make([]*Item, 0, len(ids))
	errors := make([]error, 0)

	for result := range resultCh {
		if result.Error != nil {
			errors = append(errors, fmt.Errorf("failed to get item %d: %w", result.ID, result.Error))
		} else if result.Item != nil {
			items = append(items, result.Item)
		}
	}

	// Return an error if we couldn't get any items
	if len(items) == 0 && len(errors) > 0 {
		return nil, fmt.Errorf("failed to get any items: %w", errors[0])
	}

	// Return a combined error if some items failed
	if len(errors) > 0 {
		return items, errors[0]
	}

	return items, nil
}

// itemResult holds the result of getting a single item, used by GetItemsBatch.
type itemResult struct {
	Item  *Item
	ID    int
	Error error
}
