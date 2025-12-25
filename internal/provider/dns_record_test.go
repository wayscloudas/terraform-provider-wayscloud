// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"
)

func TestNormalizeRecordType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"A", "A"},
		{"a", "A"},
		{"txt", "TXT"},
		{"TXT", "TXT"},
		{"cname", "CNAME"},
		{"CNAME", "CNAME"},
		{"MX", "MX"},
		{"mx", "MX"},
		{"Aaaa", "AAAA"},
	}

	for _, tt := range tests {
		result := normalizeRecordType(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeRecordType(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeHostname(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"www", "www"},
		{"www.", "www"},
		{"subdomain.example.com", "subdomain.example.com"},
		{"subdomain.example.com.", "subdomain.example.com"},
		{"", ""},
		{"@", "@"},
		{"*", "*"},
	}

	for _, tt := range tests {
		result := normalizeHostname(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeHostname(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeRecordValue(t *testing.T) {
	tests := []struct {
		value      string
		recordType string
		expected   string
	}{
		// CNAME - should remove trailing dot
		{"example.com", "CNAME", "example.com"},
		{"example.com.", "CNAME", "example.com"},
		{"example.com.", "cname", "example.com"},

		// MX - should remove trailing dot
		{"mail.example.com", "MX", "mail.example.com"},
		{"mail.example.com.", "MX", "mail.example.com"},

		// NS - should remove trailing dot
		{"ns1.example.com.", "NS", "ns1.example.com"},

		// SRV - should remove trailing dot
		{"_service._tcp.example.com.", "SRV", "_service._tcp.example.com"},

		// PTR - should remove trailing dot
		{"host.example.com.", "PTR", "host.example.com"},

		// A - should NOT remove trailing dot (IP address)
		{"192.0.2.1", "A", "192.0.2.1"},

		// AAAA - should NOT affect IPv6
		{"2001:db8::1", "AAAA", "2001:db8::1"},

		// TXT - should remove surrounding quotes
		{"v=spf1 include:_spf.google.com ~all", "TXT", "v=spf1 include:_spf.google.com ~all"},
		{"\"v=spf1 include:_spf.google.com ~all\"", "TXT", "v=spf1 include:_spf.google.com ~all"},
		{"\"quoted value\"", "txt", "quoted value"},

		// SPF - should also handle quotes
		{"\"v=spf1 -all\"", "SPF", "v=spf1 -all"},

		// TXT with partial quotes (should NOT change)
		{"\"only start", "TXT", "\"only start"},
		{"only end\"", "TXT", "only end\""},

		// Empty value
		{"", "A", ""},
		{"", "TXT", ""},
	}

	for _, tt := range tests {
		result := normalizeRecordValue(tt.value, tt.recordType)
		if result != tt.expected {
			t.Errorf("normalizeRecordValue(%q, %q) = %q, expected %q", tt.value, tt.recordType, result, tt.expected)
		}
	}
}
