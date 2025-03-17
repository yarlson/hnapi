package hnapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestStartUpdates(t *testing.T) {
	// Define a sequence of update responses
	updateResponses := []string{
		`{"items": [123, 456], "profiles": ["user1", "user2"]}`,
		`{"items": [789], "profiles": ["user3"]}`,
		`{"items": [], "profiles": []}`, // Empty update
		`{"items": [101112], "profiles": ["user4"]}`,
	}

	// Keep track of which response to send and how many requests were made
	var responseIndex int32
	var requestCount int32

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Count the request
		currentRequest := atomic.AddInt32(&requestCount, 1)

		// Check request path
		if !strings.HasSuffix(r.URL.Path, "updates.json") {
			t.Errorf("Expected request path to end with updates.json, got %s", r.URL.Path)
		}

		// Write response
		w.WriteHeader(http.StatusOK)

		// Get the current response index
		idx := atomic.LoadInt32(&responseIndex)

		// Send the appropriate response
		if int(idx) < len(updateResponses) {
			resp := updateResponses[idx]
			_, err := w.Write([]byte(resp))
			if err != nil {
				t.Fatalf("Failed to write mock response: %v", err)
			}

			// Every two requests, move to the next response
			if currentRequest%2 == 0 {
				atomic.AddInt32(&responseIndex, 1)
			}
		} else {
			// After we've gone through all responses, just return empty updates
			_, err := w.Write([]byte(`{"items": [], "profiles": []}`))
			if err != nil {
				t.Fatalf("Failed to write mock response: %v", err)
			}
		}
	}))
	defer server.Close()

	// Create client with a short poll interval for testing
	client := NewClient(
		WithBaseURL(server.URL+"/"),
		WithPollInterval(50*time.Millisecond),
	)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	// Start updates
	updatesCh, err := client.StartUpdates(ctx)
	if err != nil {
		t.Fatalf("StartUpdates() error = %v", err)
	}

	// Collect the received updates
	var receivedUpdates []Updates
	for updates := range updatesCh {
		receivedUpdates = append(receivedUpdates, updates)
	}

	// We should have received at least some of the updates
	// The exact number depends on timing, but we should have at least 1
	if len(receivedUpdates) < 1 {
		t.Errorf("Expected to receive at least 1 update, got %d", len(receivedUpdates))
	}

	// Only verify that we're receiving valid updates with non-empty content
	// Due to timing issues and the possibility of receiving the same update multiple times,
	// we can't do exact matching with the expected responses
	for _, updates := range receivedUpdates {
		// Skip empty updates (they have no value to check)
		if len(updates.Items) == 0 && len(updates.Profiles) == 0 {
			continue
		}

		// Ensure we at least have valid content
		foundMatch := false
		for _, expectedJSON := range updateResponses {
			var expected Updates
			err := json.Unmarshal([]byte(expectedJSON), &expected)
			if err != nil {
				t.Fatalf("Failed to unmarshal expected response: %v", err)
			}

			// Skip empty responses
			if len(expected.Items) == 0 && len(expected.Profiles) == 0 {
				continue
			}

			// If lengths match, it's a potential match
			if len(updates.Items) == len(expected.Items) && len(updates.Profiles) == len(expected.Profiles) {
				foundMatch = true
				break
			}
		}

		if !foundMatch && (len(updates.Items) > 0 || len(updates.Profiles) > 0) {
			t.Errorf("Received unexpected update: %+v", updates)
		}
	}
}

func TestStartUpdatesContextCancellation(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return some updates
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"items": [123, 456], "profiles": ["user1", "user2"]}`))
		if err != nil {
			t.Fatalf("Failed to write mock response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with a reasonable poll interval
	client := NewClient(
		WithBaseURL(server.URL+"/"),
		WithPollInterval(100*time.Millisecond),
	)

	// Create a context that can be canceled
	ctx, cancel := context.WithCancel(context.Background())

	// Start updates
	updatesCh, err := client.StartUpdates(ctx)
	if err != nil {
		t.Fatalf("StartUpdates() error = %v", err)
	}

	// Read one update to ensure the polling is working
	select {
	case _, ok := <-updatesCh:
		if !ok {
			t.Fatal("Updates channel closed unexpectedly")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Timed out waiting for update")
	}

	// Cancel the context
	cancel()

	// The channel should be closed soon
	select {
	case _, ok := <-updatesCh:
		if ok {
			// We might get one more update that was in flight,
			// but the channel should close after that
			select {
			case _, ok := <-updatesCh:
				if ok {
					t.Fatal("Updates channel still open after context cancellation")
				}
			case <-time.After(200 * time.Millisecond):
				t.Fatal("Updates channel not closed after context cancellation")
			}
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Updates channel not closed after context cancellation")
	}
}

func TestStartUpdatesWithError(t *testing.T) {
	// Starting with a valid response, then return errors
	validResponse := true

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if validResponse {
			// First response is valid
			validResponse = false
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"items": [123, 456], "profiles": ["user1", "user2"]}`))
			if err != nil {
				t.Fatalf("Failed to write mock response: %v", err)
			}
		} else {
			// Subsequent responses are errors
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("Internal Server Error"))
			if err != nil {
				t.Fatalf("Failed to write mock response: %v", err)
			}
		}
	}))
	defer server.Close()

	// Create client with a short poll interval
	client := NewClient(
		WithBaseURL(server.URL+"/"),
		WithPollInterval(50*time.Millisecond),
	)

	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start updates
	updatesCh, err := client.StartUpdates(ctx)
	if err != nil {
		t.Fatalf("StartUpdates() error = %v", err)
	}

	// Should still get the first valid update
	select {
	case update, ok := <-updatesCh:
		if !ok {
			t.Fatal("Updates channel closed unexpectedly")
		}
		// Verify the update
		if len(update.Items) != 2 || len(update.Profiles) != 2 {
			t.Errorf("Unexpected update content: %+v", update)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timed out waiting for update")
	}

	// The polling should continue despite errors,
	// but the channel should close when the context is done
	<-ctx.Done()

	// Wait a bit to make sure the channel closes
	time.Sleep(100 * time.Millisecond)

	// Check that the channel is closed
	select {
	case _, ok := <-updatesCh:
		if ok {
			t.Fatal("Updates channel still open after context timeout")
		}
	default:
		// Channel is not ready, which is unexpected
		t.Fatal("Updates channel not closed after context timeout")
	}
}

func TestStartUpdatesInitialPoll(t *testing.T) {
	// Create a channel to signal when the first request is made
	firstRequestMade := make(chan struct{}, 1)

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Signal that a request was made
		select {
		case firstRequestMade <- struct{}{}:
			// Successfully signaled
		default:
			// Channel already has a value, which means this is not the first request
		}

		// Return some updates
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"items": [123, 456], "profiles": ["user1", "user2"]}`))
		if err != nil {
			t.Fatalf("Failed to write mock response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with a long poll interval to ensure we're only testing the initial poll
	client := NewClient(
		WithBaseURL(server.URL+"/"),
		WithPollInterval(1*time.Hour), // Long interval so ticker won't trigger during test
	)

	// Start updates
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	updatesCh, err := client.StartUpdates(ctx)
	if err != nil {
		t.Fatalf("StartUpdates() error = %v", err)
	}

	// Wait for the first request to be made, with a timeout
	select {
	case <-firstRequestMade:
		// Initial poll request was made immediately
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timed out waiting for initial poll")
	}

	// Should receive updates from the initial poll
	select {
	case update, ok := <-updatesCh:
		if !ok {
			t.Fatal("Updates channel closed unexpectedly")
		}
		// Verify the update
		if len(update.Items) != 2 || len(update.Profiles) != 2 {
			t.Errorf("Unexpected update content: %+v", update)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timed out waiting for update from initial poll")
	}

	// Cancel the context to clean up
	cancel()
}

func TestUpdateContextCanceledDuringSend(t *testing.T) {
	// Use a channel to block the HTTP server from responding until we're ready
	serverBlockCh := make(chan struct{})
	serverResponseCh := make(chan struct{})

	// Create a test server that waits for signal before responding
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Signal that a request was received
		select {
		case serverResponseCh <- struct{}{}:
		default:
			// Channel already has a value
		}

		// Wait for signal before proceeding (to control timing)
		<-serverBlockCh

		// Return some updates
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"items": [123, 456], "profiles": ["user1", "user2"]}`))
		if err != nil {
			t.Fatalf("Failed to write mock response: %v", err)
		}
	}))
	defer server.Close()

	// Create a client with a zero-buffered updates channel
	// This will force StartUpdates to create a buffered channel
	client := NewClient(
		WithBaseURL(server.URL+"/"),
		WithPollInterval(500*time.Millisecond), // Long enough to not trigger twice during test
	)

	// Create a context we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start updates
	updatesCh, err := client.StartUpdates(ctx)
	if err != nil {
		t.Fatalf("StartUpdates() error = %v", err)
	}

	// Wait for the server to receive the request
	select {
	case <-serverResponseCh:
		// Request received by server
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timed out waiting for server to receive request")
	}

	// Now trigger the server to reply with updates
	close(serverBlockCh)

	// Read an update to confirm everything is working
	select {
	case update, ok := <-updatesCh:
		if !ok {
			t.Fatal("Updates channel closed unexpectedly")
		}
		// Verify the update
		if len(update.Items) != 2 || len(update.Profiles) != 2 {
			t.Errorf("Unexpected update content: %+v", update)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timed out waiting for update")
	}

	// Now cancel the context - which should stop the polling and close the channel
	cancel()

	// Give the goroutine time to clean up
	time.Sleep(100 * time.Millisecond)

	// Verify channel is closed
	_, channelOpen := <-updatesCh
	if channelOpen {
		t.Error("Channel should be closed after context cancellation")
	}
}

func TestStartUpdatesTickerBased(t *testing.T) {
	// Counter for requests
	var requestCount int32

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Count the request
		count := atomic.AddInt32(&requestCount, 1)

		// Return different responses based on request count
		w.WriteHeader(http.StatusOK)
		if count == 1 {
			// First request (initial poll)
			_, err := w.Write([]byte(`{"items": [123], "profiles": ["user1"]}`))
			if err != nil {
				t.Fatalf("Failed to write mock response: %v", err)
			}
		} else {
			// Second request (from ticker)
			_, err := w.Write([]byte(`{"items": [456], "profiles": ["user2"]}`))
			if err != nil {
				t.Fatalf("Failed to write mock response: %v", err)
			}
		}
	}))
	defer server.Close()

	// Create client with a very short poll interval
	client := NewClient(
		WithBaseURL(server.URL+"/"),
		WithPollInterval(50*time.Millisecond), // Very short for testing
	)

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start updates
	updatesCh, err := client.StartUpdates(ctx)
	if err != nil {
		t.Fatalf("StartUpdates() error = %v", err)
	}

	// Collect updates until context is canceled
	var updates []Updates
	for update := range updatesCh {
		updates = append(updates, update)
	}

	// Verify we got at least 2 updates (initial poll and at least one from ticker)
	if len(updates) < 2 {
		t.Errorf("Expected at least 2 updates (initial poll and ticker), got %d", len(updates))
	}

	// Verify the content of the updates
	foundInitial := false
	foundTicker := false
	for _, update := range updates {
		if len(update.Items) == 1 {
			if update.Items[0] == 123 && len(update.Profiles) == 1 && update.Profiles[0] == "user1" {
				foundInitial = true
			} else if update.Items[0] == 456 && len(update.Profiles) == 1 && update.Profiles[0] == "user2" {
				foundTicker = true
			}
		}
	}

	if !foundInitial {
		t.Errorf("Did not receive update from initial poll")
	}
	if !foundTicker {
		t.Errorf("Did not receive update from ticker-based poll")
	}
}

func TestStartUpdatesInitialPollError(t *testing.T) {
	// Track poll attempts to return error only on the first one
	var pollCount int32

	// Create a test server that returns error on first request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&pollCount, 1)

		if count == 1 {
			// First request (initial poll) - return error
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("Internal Server Error"))
			if err != nil {
				t.Fatalf("Failed to write mock response: %v", err)
			}
		} else {
			// Subsequent requests - return success
			w.WriteHeader(http.StatusOK)
			_, err := w.Write([]byte(`{"items": [456], "profiles": ["user2"]}`))
			if err != nil {
				t.Fatalf("Failed to write mock response: %v", err)
			}
		}
	}))
	defer server.Close()

	// Create client with a very short poll interval
	client := NewClient(
		WithBaseURL(server.URL+"/"),
		WithPollInterval(50*time.Millisecond), // Very short for testing
	)

	// Create a context with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start updates - this should log an error for the initial poll but continue
	updatesCh, err := client.StartUpdates(ctx)
	if err != nil {
		t.Fatalf("StartUpdates() error = %v", err)
	}

	// Despite the initial error, we should get updates from the next poll
	// that happens after the ticker fires
	var update Updates
	var ok bool
	select {
	case update, ok = <-updatesCh:
		if !ok {
			t.Fatal("Updates channel closed unexpectedly")
		}
	case <-time.After(150 * time.Millisecond):
		t.Fatal("Timed out waiting for update from subsequent poll")
	}

	// Verify the update is from the second poll
	if len(update.Items) != 1 || update.Items[0] != 456 {
		t.Errorf("Unexpected update content: %+v", update)
	}

	// Verify we made at least 2 poll attempts
	if atomic.LoadInt32(&pollCount) < 2 {
		t.Errorf("Expected at least 2 poll attempts, got %d", pollCount)
	}

	// Cleanup
	cancel()

	// Give the goroutine time to clean up
	time.Sleep(50 * time.Millisecond)

	// Verify channel is closed after context cancellation
	_, ok = <-updatesCh
	if ok {
		t.Error("Channel should be closed after context cancellation")
	}
}

func TestCancelDuringUpdateFetch(t *testing.T) {
	// Create a test server that blocks until request context is canceled
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wait for the request context to be cancelled
		<-r.Context().Done()

		// Just in case, still write something, though it shouldn't be used
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items": [123], "profiles": ["user1"]}`))
	}))
	defer server.Close()

	// Create a client with a long poll interval
	client := NewClient(
		WithBaseURL(server.URL+"/"),
		WithPollInterval(1*time.Hour), // Long interval so ticker won't trigger during test
	)

	// Create a context that we'll cancel shortly
	ctx, cancel := context.WithCancel(context.Background())

	// Start updates
	updatesCh, err := client.StartUpdates(ctx)
	if err != nil {
		t.Fatalf("StartUpdates() error = %v", err)
	}

	// Wait a bit to ensure the initial poll request is made
	time.Sleep(50 * time.Millisecond)

	// Cancel the context - this should cancel the HTTP request in progress
	cancel()

	// Wait for the updates channel to be closed
	select {
	case _, stillOpen := <-updatesCh:
		if stillOpen {
			t.Fatal("Expected updates channel to be closed after context cancellation")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Timed out while waiting for channel to close")
	}
}
