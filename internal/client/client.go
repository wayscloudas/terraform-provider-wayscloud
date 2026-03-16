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
	"strings"
	"time"
)

const (
	DefaultBaseURL = "https://api.wayscloud.services"
	DefaultTimeout = 30 * time.Second
	MaxRetries     = 3
	RetryWaitTime  = 1 * time.Second
)

// Client is the WAYSCloud API client
type Client struct {
	BaseURL    string
	APIKey     string
	PATToken   string // Personal Access Token (for database API)
	HTTPClient *http.Client
	UserAgent  string
}

// NewClient creates a new WAYSCloud API client (legacy single-token mode)
func NewClient(apiKey string, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	c := &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		UserAgent: "terraform-provider-wayscloud",
	}

	// Detect auth type from key format
	// PAT tokens start with "wayscloud_pat_"
	// API keys start with "wayscloud_api_"
	if strings.HasPrefix(apiKey, "wayscloud_pat_") {
		c.PATToken = apiKey
	} else {
		c.APIKey = apiKey
	}

	return c
}

// NewClientDualAuth creates a client that supports both API key and PAT authentication.
// API key is used via X-API-Key header for most resources.
// PAT is used via Authorization: Bearer for database and domain verification resources.
// If only one token is provided, it auto-detects the type (backward compatible).
func NewClientDualAuth(apiKey string, patToken string, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	c := &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: DefaultTimeout,
		},
		UserAgent: "terraform-provider-wayscloud",
	}

	// If explicit PAT token provided, use it
	if patToken != "" {
		c.PATToken = patToken
	}

	// If explicit API key provided, use it
	if apiKey != "" {
		// If apiKey is actually a PAT and no separate PAT was provided, use it as PAT too
		if strings.HasPrefix(apiKey, "wayscloud_pat_") {
			if c.PATToken == "" {
				c.PATToken = apiKey
			}
			// Don't set as APIKey since PAT uses Bearer auth
		} else {
			c.APIKey = apiKey
		}
	}

	return c
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

// isRetryable returns true if the status code indicates a transient error
func isRetryable(statusCode int) bool {
	return statusCode == 429 || statusCode == 502 || statusCode == 503 || statusCode == 504
}

// sanitizeErrorMessage removes potentially sensitive content from error messages
// while preserving useful debugging information
func sanitizeErrorMessage(msg string) string {
	// Truncate very long messages that might contain debug info
	if len(msg) > 500 {
		msg = msg[:500] + "..."
	}
	// Mask API keys (wayscloud_api_xxx...) but keep context
	// Pattern: wayscloud_api_ followed by 10-char prefix, underscore, 56-char secret
	msg = maskTokenPattern(msg, "wayscloud_api_")
	// Mask PAT tokens (wayscloud_pat_xxx...) but keep context
	msg = maskTokenPattern(msg, "wayscloud_pat_")
	return msg
}

// maskTokenPattern replaces tokens matching the prefix with a masked version
// e.g., "wayscloud_api_abc1234567_secret..." → "[REDACTED]"
func maskTokenPattern(msg, prefix string) string {
	result := msg
	searchStart := 0
	for {
		idx := strings.Index(result[searchStart:], prefix)
		if idx == -1 {
			break
		}
		idx += searchStart // Adjust for the offset

		// Find the end of the token (space, quote, or end of string)
		endIdx := idx + len(prefix)
		for endIdx < len(result) {
			c := result[endIdx]
			if c == ' ' || c == '"' || c == '\'' || c == ',' || c == '}' || c == '\n' || c == '\r' {
				break
			}
			endIdx++
		}
		// Replace the entire token with [REDACTED]
		replacement := "[REDACTED]"
		result = result[:idx] + replacement + result[endIdx:]
		// Move search start past the replacement to avoid infinite loop
		searchStart = idx + len(replacement)
	}
	return result
}

// doRequest performs an HTTP request with authentication and retry logic
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var jsonBody []byte
	var err error

	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt < MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(RetryWaitTime * time.Duration(attempt)):
				// Exponential backoff
			}
		}

		var bodyReader io.Reader
		if jsonBody != nil {
			bodyReader = bytes.NewReader(jsonBody)
		}

		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set auth headers - both can be present for dual-auth support
		if c.APIKey != "" {
			req.Header.Set("X-API-Key", c.APIKey)
		}
		if c.PATToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.PATToken)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", c.UserAgent)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to execute request: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		// Check for retryable errors
		if isRetryable(resp.StatusCode) {
			lastErr = &APIError{
				StatusCode: resp.StatusCode,
				Message:    fmt.Sprintf("transient error (attempt %d/%d)", attempt+1, MaxRetries),
			}
			continue
		}

		if resp.StatusCode >= 400 {
			var apiErr struct {
				Detail  string `json:"detail"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(respBody, &apiErr); err == nil {
				return nil, &APIError{
					StatusCode: resp.StatusCode,
					Message:    sanitizeErrorMessage(apiErr.Message),
					Detail:     sanitizeErrorMessage(apiErr.Detail),
				}
			}
			// Fallback: sanitize raw response
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Message:    sanitizeErrorMessage(string(respBody)),
			}
		}

		return respBody, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("request failed after %d attempts", MaxRetries)
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
