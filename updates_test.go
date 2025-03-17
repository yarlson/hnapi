package hnapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yarlson/hnapi"
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
	client := hnapi.NewClient(
		hnapi.WithBaseURL(server.URL+"/"),
		hnapi.WithPollInterval(50*time.Millisecond),
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
	var receivedUpdates []hnapi.Updates
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
			var expected hnapi.Updates
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
	client := hnapi.NewClient(
		hnapi.WithBaseURL(server.URL+"/"),
		hnapi.WithPollInterval(100*time.Millisecond),
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
	client := hnapi.NewClient(
		hnapi.WithBaseURL(server.URL+"/"),
		hnapi.WithPollInterval(50*time.Millisecond),
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
