package hnapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
)

// GetItem retrieves a single Hacker News item by its ID.
// It returns the item or an error if the request fails or the context is canceled.
func (c *Client) GetItem(ctx context.Context, id int) (*Item, error) {
	// Construct the URL for the item endpoint
	endpoint := path.Join("item", fmt.Sprintf("%d.json", id))

	// Make the request
	var item Item
	if err := c.makeRequest(ctx, endpoint, &item); err != nil {
		return nil, fmt.Errorf("failed to get item %d: %w", id, err)
	}

	return &item, nil
}

// GetUser retrieves a Hacker News user by username.
// It returns the user or an error if the request fails or the context is canceled.
func (c *Client) GetUser(ctx context.Context, username string) (*User, error) {
	// Construct the URL for the user endpoint
	endpoint := path.Join("user", fmt.Sprintf("%s.json", username))

	// Make the request
	var user User
	if err := c.makeRequest(ctx, endpoint, &user); err != nil {
		return nil, fmt.Errorf("failed to get user %s: %w", username, err)
	}

	return &user, nil
}

// makeRequest performs an HTTP GET request to the specified endpoint and unmarshals the response into the target.
// It uses the client's configuration for the base URL and timeout.
func (c *Client) makeRequest(ctx context.Context, endpoint string, target interface{}) error {
	// Create a new HTTP request
	fullURL := c.Config.BaseURL + endpoint

	// Create a new request with the provided context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Execute the request
	resp, err := c.Config.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read and parse the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// If we got an empty response or "null", return an error
	if len(body) == 0 || string(body) == "null" {
		return fmt.Errorf("item not found or null response")
	}

	// Unmarshal the JSON response into the target
	if err := json.Unmarshal(body, target); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}
