// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"testing"
)

func TestNewClient_APIKey(t *testing.T) {
	c := NewClient("wayscloud_api_prefix1234_secretsecretsecretsecretsecretsecretsecretsecret", "")

	if c.APIKey == "" {
		t.Error("Expected APIKey to be set")
	}
	if c.PATToken != "" {
		t.Error("Expected PATToken to be empty for API key")
	}
	if c.BaseURL != DefaultBaseURL {
		t.Errorf("Expected BaseURL %s, got %s", DefaultBaseURL, c.BaseURL)
	}
}

func TestNewClient_PAT(t *testing.T) {
	c := NewClient("wayscloud_pat_prefix1234_secretsecretsecretsecretsecretsecretsecretsecret", "")

	if c.PATToken == "" {
		t.Error("Expected PATToken to be set")
	}
	if c.APIKey != "" {
		t.Error("Expected APIKey to be empty for PAT token")
	}
}

func TestNewClient_CustomBaseURL(t *testing.T) {
	c := NewClient("wayscloud_api_test", "https://api-staging.wayscloud.services")

	if c.BaseURL != "https://api-staging.wayscloud.services" {
		t.Errorf("Expected custom BaseURL, got %s", c.BaseURL)
	}
}

func TestSanitizeErrorMessage_MasksAPIKey(t *testing.T) {
	msg := "Authentication failed for wayscloud_api_abc1234567_verylongsecretkey123456"
	sanitized := sanitizeErrorMessage(msg)

	if sanitized == msg {
		t.Error("Expected message to be sanitized")
	}
	if sanitized != "Authentication failed for [REDACTED]" {
		t.Errorf("Unexpected sanitized message: %s", sanitized)
	}
}

func TestSanitizeErrorMessage_MasksPAT(t *testing.T) {
	msg := "Token wayscloud_pat_xyz9876543_anothersecretkey789 is invalid"
	sanitized := sanitizeErrorMessage(msg)

	if sanitized != "Token [REDACTED] is invalid" {
		t.Errorf("Unexpected sanitized message: %s", sanitized)
	}
}

func TestSanitizeErrorMessage_MasksMultipleTokens(t *testing.T) {
	msg := "API: wayscloud_api_key1 PAT: wayscloud_pat_token1"
	sanitized := sanitizeErrorMessage(msg)

	expected := "API: [REDACTED] PAT: [REDACTED]"
	if sanitized != expected {
		t.Errorf("Expected %s, got %s", expected, sanitized)
	}
}

func TestSanitizeErrorMessage_PreservesNonTokenContent(t *testing.T) {
	msg := "HTTP 401: Invalid credentials - check your API key configuration"
	sanitized := sanitizeErrorMessage(msg)

	if sanitized != msg {
		t.Errorf("Expected message to be unchanged, got: %s", sanitized)
	}
}

func TestSanitizeErrorMessage_TruncatesLongMessages(t *testing.T) {
	// Create a 600 character message
	longMsg := ""
	for i := 0; i < 600; i++ {
		longMsg += "x"
	}

	sanitized := sanitizeErrorMessage(longMsg)

	if len(sanitized) != 503 { // 500 + "..."
		t.Errorf("Expected truncated length 503, got %d", len(sanitized))
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{200, false},
		{400, false},
		{401, false},
		{404, false},
		{429, true},  // Rate limit
		{500, false},
		{502, true},  // Bad gateway
		{503, true},  // Service unavailable
		{504, true},  // Gateway timeout
	}

	for _, tt := range tests {
		result := isRetryable(tt.code)
		if result != tt.expected {
			t.Errorf("isRetryable(%d) = %v, expected %v", tt.code, result, tt.expected)
		}
	}
}
