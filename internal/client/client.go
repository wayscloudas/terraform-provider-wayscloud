// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultBaseURL = "https://api.wayscloud.services"
	DefaultTimeout = 30 * time.Second
)

// Client is the WAYSCloud API client
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
	UserAgent  string
}

// NewClient creates a new WAYSCloud API client
func NewClient(apiKey string, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		UserAgent: "terraform-provider-wayscloud",
	}
}

// APIError represents an error returned by the API
type APIError struct {
	StatusCode int
	Message    string
	Detail     string
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("API error %d: %s - %s", e.StatusCode, e.Message, e.Detail)
	}
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}

// doRequest performs an HTTP request with authentication
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Detail  string `json:"detail"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(respBody, &apiErr); err == nil {
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Message:    apiErr.Message,
				Detail:     apiErr.Detail,
			}
		}
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	return respBody, nil
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPost, path, body)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPut, path, body)
}

// Patch performs a PATCH request
func (c *Client) Patch(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPatch, path, body)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil)
}
