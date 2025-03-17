package hnapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yarlson/hnapi"
)

func TestGetItem(t *testing.T) {
	// Set up test cases
	tests := []struct {
		name           string
		id             int
		mockResponse   string
		mockStatusCode int
		wantErr        bool
		validateItem   func(*testing.T, *hnapi.Item)
	}{
		{
			name:           "valid story item",
			id:             8863,
			mockStatusCode: http.StatusOK,
			mockResponse: `{
				"by": "dhouston",
				"descendants": 71,
				"id": 8863,
				"kids": [8952, 9224, 8917],
				"score": 111,
				"time": 1175714200,
				"title": "My YC app: Dropbox - Throw away your USB drive",
				"type": "story",
				"url": "http://www.getdropbox.com/u/2/screencast.html"
			}`,
			wantErr: false,
			validateItem: func(t *testing.T, item *hnapi.Item) {
				if item.ID != 8863 {
					t.Errorf("Expected ID to be 8863, got %d", item.ID)
				}
				if item.Type != "story" {
					t.Errorf("Expected Type to be 'story', got %q", item.Type)
				}
				if item.By != "dhouston" {
					t.Errorf("Expected By to be 'dhouston', got %q", item.By)
				}
			},
		},
		{
			name:           "item not found",
			id:             999999,
			mockStatusCode: http.StatusOK,
			mockResponse:   "null",
			wantErr:        true,
			validateItem:   nil,
		},
		{
			name:           "server error",
			id:             8863,
			mockStatusCode: http.StatusInternalServerError,
			mockResponse:   "Internal Server Error",
			wantErr:        true,
			validateItem:   nil,
		},
		{
			name:           "invalid json",
			id:             8863,
			mockStatusCode: http.StatusOK,
			mockResponse:   "{invalid json",
			wantErr:        true,
			validateItem:   nil,
		},
		{
			name:           "empty response",
			id:             8863,
			mockStatusCode: http.StatusOK,
			mockResponse:   "",
			wantErr:        true,
			validateItem:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request path
				expectedPath := fmt.Sprintf("/item/%d.json", tt.id)
				if !strings.HasSuffix(r.URL.Path, expectedPath) {
					t.Errorf("Expected request path to end with %s, got %s", expectedPath, r.URL.Path)
				}

				// Set status code and write response
				w.WriteHeader(tt.mockStatusCode)
				_, err := w.Write([]byte(tt.mockResponse))
				if err != nil {
					t.Fatalf("Failed to write mock response: %v", err)
				}
			}))
			defer server.Close()

			// Create client with the test server URL
			client := hnapi.NewClient(hnapi.WithBaseURL(server.URL + "/"))

			// Call GetItem
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			item, err := client.GetItem(ctx, tt.id)

			// Check if error matches expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("GetItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Validate item if no error
			if err == nil && tt.validateItem != nil {
				tt.validateItem(t, item)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	// Set up test cases
	tests := []struct {
		name           string
		username       string
		mockResponse   string
		mockStatusCode int
		wantErr        bool
		validateUser   func(*testing.T, *hnapi.User)
	}{
		{
			name:           "valid user",
			username:       "jl",
			mockStatusCode: http.StatusOK,
			mockResponse: `{
				"about": "This is a test",
				"created": 1173923446,
				"id": "jl",
				"karma": 2937,
				"submitted": [8265435, 8168423, 8090946]
			}`,
			wantErr: false,
			validateUser: func(t *testing.T, user *hnapi.User) {
				if user.ID != "jl" {
					t.Errorf("Expected ID to be 'jl', got %q", user.ID)
				}
				if user.Karma != 2937 {
					t.Errorf("Expected Karma to be 2937, got %d", user.Karma)
				}
				if user.Created != 1173923446 {
					t.Errorf("Expected Created to be 1173923446, got %d", user.Created)
				}
			},
		},
		{
			name:           "user not found",
			username:       "nonexistentuser",
			mockStatusCode: http.StatusOK,
			mockResponse:   "null",
			wantErr:        true,
			validateUser:   nil,
		},
		{
			name:           "server error",
			username:       "jl",
			mockStatusCode: http.StatusInternalServerError,
			mockResponse:   "Internal Server Error",
			wantErr:        true,
			validateUser:   nil,
		},
		{
			name:           "invalid json",
			username:       "jl",
			mockStatusCode: http.StatusOK,
			mockResponse:   "{invalid json",
			wantErr:        true,
			validateUser:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request path
				expectedPath := fmt.Sprintf("/user/%s.json", tt.username)
				if !strings.HasSuffix(r.URL.Path, expectedPath) {
					t.Errorf("Expected request path to end with %s, got %s", expectedPath, r.URL.Path)
				}

				// Set status code and write response
				w.WriteHeader(tt.mockStatusCode)
				_, err := w.Write([]byte(tt.mockResponse))
				if err != nil {
					t.Fatalf("Failed to write mock response: %v", err)
				}
			}))
			defer server.Close()

			// Create client with the test server URL
			client := hnapi.NewClient(hnapi.WithBaseURL(server.URL + "/"))

			// Call GetUser
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			user, err := client.GetUser(ctx, tt.username)

			// Check if error matches expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Validate user if no error
			if err == nil && tt.validateUser != nil {
				tt.validateUser(t, user)
			}
		})
	}
}

func TestContextCancellation(t *testing.T) {
	// Create a server that delays its response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay the response
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		item := &hnapi.Item{ID: 8863, Type: "story"}
		json.NewEncoder(w).Encode(item)
	}))
	defer server.Close()

	// Create client with the test server URL
	client := hnapi.NewClient(hnapi.WithBaseURL(server.URL + "/"))

	// Create a context that will be canceled immediately
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Call GetItem with a context that will time out
	_, err := client.GetItem(ctx, 8863)

	// Expect a context deadline exceeded error
	if err == nil {
		t.Errorf("Expected error due to context cancellation, got nil")
	}
}
