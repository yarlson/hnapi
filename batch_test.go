package hnapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetItemsBatch(t *testing.T) {
	tests := []struct {
		name              string
		ids               []int
		concurrency       int
		maxConcurrent     int32 // To track maximum concurrent requests
		mockResponses     map[int]string
		mockStatusCodes   map[int]int
		responseDelay     time.Duration // To simulate slower responses
		expectedItemCount int
		wantErr           bool
	}{
		{
			name:              "successful batch",
			ids:               []int{8863, 8864, 8865},
			concurrency:       2,
			maxConcurrent:     0,
			responseDelay:     10 * time.Millisecond,
			expectedItemCount: 3,
			mockResponses: map[int]string{
				8863: `{"id": 8863, "type": "story", "title": "Test Story 1", "by": "user1"}`,
				8864: `{"id": 8864, "type": "story", "title": "Test Story 2", "by": "user2"}`,
				8865: `{"id": 8865, "type": "story", "title": "Test Story 3", "by": "user3"}`,
			},
			mockStatusCodes: map[int]int{
				8863: http.StatusOK,
				8864: http.StatusOK,
				8865: http.StatusOK,
			},
			wantErr: false,
		},
		{
			name:              "partial success",
			ids:               []int{8863, 8864, 8865},
			concurrency:       2,
			maxConcurrent:     0,
			responseDelay:     10 * time.Millisecond,
			expectedItemCount: 2,
			mockResponses: map[int]string{
				8863: `{"id": 8863, "type": "story", "title": "Test Story 1", "by": "user1"}`,
				8864: `{"id": 8864, "type": "story", "title": "Test Story 2", "by": "user2"}`,
				8865: `null`, // This will cause an error
			},
			mockStatusCodes: map[int]int{
				8863: http.StatusOK,
				8864: http.StatusOK,
				8865: http.StatusOK,
			},
			wantErr: true,
		},
		{
			name:              "all failures",
			ids:               []int{8863, 8864, 8865},
			concurrency:       2,
			maxConcurrent:     0,
			responseDelay:     10 * time.Millisecond,
			expectedItemCount: 0,
			mockResponses: map[int]string{
				8863: `null`,
				8864: `null`,
				8865: `null`,
			},
			mockStatusCodes: map[int]int{
				8863: http.StatusOK,
				8864: http.StatusOK,
				8865: http.StatusOK,
			},
			wantErr: true,
		},
		{
			name:              "empty ids",
			ids:               []int{},
			concurrency:       2,
			maxConcurrent:     0,
			responseDelay:     0,
			expectedItemCount: 0,
			mockResponses:     map[int]string{},
			mockStatusCodes:   map[int]int{},
			wantErr:           false,
		},
		{
			name:              "respect concurrency limit",
			ids:               []int{8863, 8864, 8865, 8866, 8867, 8868},
			concurrency:       3,
			maxConcurrent:     0,
			responseDelay:     50 * time.Millisecond, // Longer delay to ensure overlap
			expectedItemCount: 6,
			mockResponses: map[int]string{
				8863: `{"id": 8863, "type": "story", "title": "Test Story 1", "by": "user1"}`,
				8864: `{"id": 8864, "type": "story", "title": "Test Story 2", "by": "user2"}`,
				8865: `{"id": 8865, "type": "story", "title": "Test Story 3", "by": "user3"}`,
				8866: `{"id": 8866, "type": "story", "title": "Test Story 4", "by": "user4"}`,
				8867: `{"id": 8867, "type": "story", "title": "Test Story 5", "by": "user5"}`,
				8868: `{"id": 8868, "type": "story", "title": "Test Story 6", "by": "user6"}`,
			},
			mockStatusCodes: map[int]int{
				8863: http.StatusOK,
				8864: http.StatusOK,
				8865: http.StatusOK,
				8866: http.StatusOK,
				8867: http.StatusOK,
				8868: http.StatusOK,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Track concurrent requests
			var currentConcurrent int32
			var mutex sync.Mutex
			requestCount := make(map[int]int)

			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Extract the item ID from the path
				parts := strings.Split(r.URL.Path, "/")
				idStr := strings.TrimSuffix(parts[len(parts)-1], ".json")
				var id int
				_, err := fmt.Sscanf(idStr, "%d", &id)
				if err != nil {
					t.Errorf("Failed to parse ID from path: %v", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				// Track request counts for this ID
				mutex.Lock()
				requestCount[id]++
				mutex.Unlock()

				// Update concurrent count
				current := atomic.AddInt32(&currentConcurrent, 1)
				if current > tt.maxConcurrent {
					atomic.StoreInt32(&tt.maxConcurrent, current)
				}
				defer atomic.AddInt32(&currentConcurrent, -1)

				// Simulate work/delay
				if tt.responseDelay > 0 {
					time.Sleep(tt.responseDelay)
				}

				// Return the mock response for this ID
				if statusCode, ok := tt.mockStatusCodes[id]; ok {
					w.WriteHeader(statusCode)
				} else {
					w.WriteHeader(http.StatusNotFound)
				}

				if response, ok := tt.mockResponses[id]; ok {
					_, err := w.Write([]byte(response))
					if err != nil {
						t.Fatalf("Failed to write mock response: %v", err)
					}
				} else {
					_, err := w.Write([]byte("null"))
					if err != nil {
						t.Fatalf("Failed to write mock response: %v", err)
					}
				}
			}))
			defer server.Close()

			// Create client with custom concurrency limit
			client := NewClient(
				WithBaseURL(server.URL+"/"),
				WithConcurrency(tt.concurrency),
			)

			// Call GetItemsBatch
			ctx := context.Background()
			items, err := client.GetItemsBatch(ctx, tt.ids)

			// Check error expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("GetItemsBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check number of items returned
			if len(items) != tt.expectedItemCount {
				t.Errorf("GetItemsBatch() returned %d items, expected %d", len(items), tt.expectedItemCount)
			}

			// Verify we respected the concurrency limit
			if int(tt.maxConcurrent) > tt.concurrency {
				t.Errorf("GetItemsBatch() exceeded concurrency limit, max concurrent requests: %d, limit: %d",
					tt.maxConcurrent, tt.concurrency)
			}

			// Verify each ID was requested exactly once
			for _, id := range tt.ids {
				if count, ok := requestCount[id]; !ok || count != 1 {
					t.Errorf("ID %d was requested %d times, expected 1", id, count)
				}
			}
		})
	}
}

func TestGetItemsBatchWithContext(t *testing.T) {
	// Create a test server that delays responses
	var requestCounter int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Count the request
		atomic.AddInt32(&requestCounter, 1)

		// Delay the response significantly
		time.Sleep(200 * time.Millisecond)

		// Return a valid item
		w.WriteHeader(http.StatusOK)
		item := map[string]interface{}{
			"id":    8863,
			"type":  "story",
			"title": "Test Story",
		}
		if err := json.NewEncoder(w).Encode(item); err != nil {
			t.Fatalf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client
	client := NewClient(
		WithBaseURL(server.URL+"/"),
		WithConcurrency(5),
	)

	// Create a context that will be canceled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Try to get multiple items, but the context will be canceled before they complete
	_, err := client.GetItemsBatch(ctx, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	// We should get an error
	if err == nil {
		t.Errorf("Expected error due to context cancellation, got nil")
	}

	// Wait a moment to allow for any in-flight requests to complete
	time.Sleep(300 * time.Millisecond)

	// Check that not all requests were made (context should have canceled some)
	requestsMade := atomic.LoadInt32(&requestCounter)
	if requestsMade >= 10 {
		t.Errorf("Expected fewer than 10 requests due to context cancellation, got %d", requestsMade)
	}
}
