package hnapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestGetStories(t *testing.T) {
	// Define test cases
	tests := []struct {
		name               string
		endpoint           string
		mockResponse       string
		mockStatusCode     int
		wantErr            bool
		expectedStoryCount int
		clientMethod       func(*Client, context.Context) ([]int, error)
	}{
		{
			name:               "TopStories",
			endpoint:           "topstories.json",
			mockResponse:       `[8863, 8864, 8865, 8866, 8867]`,
			mockStatusCode:     http.StatusOK,
			wantErr:            false,
			expectedStoryCount: 5,
			clientMethod:       func(c *Client, ctx context.Context) ([]int, error) { return c.GetTopStories(ctx) },
		},
		{
			name:               "NewStories",
			endpoint:           "newstories.json",
			mockResponse:       `[9873, 9874, 9875, 9876]`,
			mockStatusCode:     http.StatusOK,
			wantErr:            false,
			expectedStoryCount: 4,
			clientMethod:       func(c *Client, ctx context.Context) ([]int, error) { return c.GetNewStories(ctx) },
		},
		{
			name:               "BestStories",
			endpoint:           "beststories.json",
			mockResponse:       `[7111, 7112, 7113]`,
			mockStatusCode:     http.StatusOK,
			wantErr:            false,
			expectedStoryCount: 3,
			clientMethod:       func(c *Client, ctx context.Context) ([]int, error) { return c.GetBestStories(ctx) },
		},
		{
			name:               "AskStories",
			endpoint:           "askstories.json",
			mockResponse:       `[6541, 6542]`,
			mockStatusCode:     http.StatusOK,
			wantErr:            false,
			expectedStoryCount: 2,
			clientMethod:       func(c *Client, ctx context.Context) ([]int, error) { return c.GetAskStories(ctx) },
		},
		{
			name:               "ShowStories",
			endpoint:           "showstories.json",
			mockResponse:       `[5321, 5322, 5323, 5324]`,
			mockStatusCode:     http.StatusOK,
			wantErr:            false,
			expectedStoryCount: 4,
			clientMethod:       func(c *Client, ctx context.Context) ([]int, error) { return c.GetShowStories(ctx) },
		},
		{
			name:               "JobStories",
			endpoint:           "jobstories.json",
			mockResponse:       `[4241, 4242]`,
			mockStatusCode:     http.StatusOK,
			wantErr:            false,
			expectedStoryCount: 2,
			clientMethod:       func(c *Client, ctx context.Context) ([]int, error) { return c.GetJobStories(ctx) },
		},
		{
			name:               "EmptyList",
			endpoint:           "topstories.json",
			mockResponse:       `[]`,
			mockStatusCode:     http.StatusOK,
			wantErr:            false,
			expectedStoryCount: 0,
			clientMethod:       func(c *Client, ctx context.Context) ([]int, error) { return c.GetTopStories(ctx) },
		},
		{
			name:               "ServerError",
			endpoint:           "topstories.json",
			mockResponse:       `Internal Server Error`,
			mockStatusCode:     http.StatusInternalServerError,
			wantErr:            true,
			expectedStoryCount: 0,
			clientMethod:       func(c *Client, ctx context.Context) ([]int, error) { return c.GetTopStories(ctx) },
		},
		{
			name:               "InvalidJSON",
			endpoint:           "topstories.json",
			mockResponse:       `{invalid json`,
			mockStatusCode:     http.StatusOK,
			wantErr:            true,
			expectedStoryCount: 0,
			clientMethod:       func(c *Client, ctx context.Context) ([]int, error) { return c.GetTopStories(ctx) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check request path
				if !strings.HasSuffix(r.URL.Path, tt.endpoint) {
					t.Errorf("Expected request path to end with %s, got %s", tt.endpoint, r.URL.Path)
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
			client := NewClient(WithBaseURL(server.URL + "/"))

			// Set up context
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Call the appropriate method
			stories, err := tt.clientMethod(client, ctx)

			// Check if error matches expectation
			if (err != nil) != tt.wantErr {
				t.Errorf("Error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// If no error, verify the result
			if err == nil {
				// Verify the length of the returned list
				if len(stories) != tt.expectedStoryCount {
					t.Errorf("Expected %d stories, got %d", tt.expectedStoryCount, len(stories))
				}

				// For success cases with non-empty lists, verify the IDs
				if tt.mockStatusCode == http.StatusOK && tt.expectedStoryCount > 0 {
					var expected []int
					err := json.Unmarshal([]byte(tt.mockResponse), &expected)
					if err != nil {
						t.Fatalf("Failed to unmarshal expected response: %v", err)
					}

					if !reflect.DeepEqual(stories, expected) {
						t.Errorf("Expected stories %v, got %v", expected, stories)
					}
				}
			}
		})
	}
}
