// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// Operation types for diagnostics
const (
	opCreate = "create"
	opRead   = "read"
	opUpdate = "update"
	opDelete = "delete"
)

// is404 returns true if the error is an API 404 Not Found
func is404(err error) bool {
	if apiErr, ok := err.(*client.APIError); ok {
		return apiErr.StatusCode == 404
	}
	return false
}

// apiDiagnostic creates structured diagnostics from API errors
func apiDiagnostic(op, resourceType string, err error) diag.Diagnostics {
	var diags diag.Diagnostics

	if apiErr, ok := err.(*client.APIError); ok {
		summary := fmt.Sprintf("WAYSCloud API Error (%s %s)", op, resourceType)
		detail := fmt.Sprintf("HTTP %d: %s", apiErr.StatusCode, apiErr.Message)
		if apiErr.Detail != "" {
			detail += "\n\nDetail: " + apiErr.Detail
		}

		switch {
		case apiErr.StatusCode == 401 || apiErr.StatusCode == 403:
			detail += "\n\nCheck that your API key or PAT has the required permissions for this resource."
		case apiErr.StatusCode == 409:
			detail += "\n\nA resource with the same identifier may already exist."
		case apiErr.StatusCode == 422:
			detail += "\n\nThe request was well-formed but contained invalid parameters."
		case apiErr.StatusCode == 429:
			detail += "\n\nRate limit exceeded. Try again later or reduce parallelism with -parallelism=1."
		case apiErr.StatusCode >= 500:
			detail += "\n\nThis is a server-side error. If it persists, contact WAYSCloud support."
		}

		diags.AddError(summary, detail)
	} else {
		// Non-API error (network, DNS, TLS, etc.)
		summary := fmt.Sprintf("WAYSCloud Client Error (%s %s)", op, resourceType)
		detail := classifyNonAPIError(err)
		diags.AddError(summary, detail)
	}

	return diags
}

// dataSourceDiagnostic creates diagnostics for data source read errors
func dataSourceDiagnostic(dataSourceType string, err error) diag.Diagnostics {
	return apiDiagnostic(opRead, dataSourceType, err)
}

// classifyNonAPIError provides user-friendly messages for common network errors
func classifyNonAPIError(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "context deadline exceeded"):
		return "Connection to WAYSCloud API timed out. Check network connectivity and try again."
	case strings.Contains(msg, "no such host"):
		return fmt.Sprintf("Unable to resolve API hostname. Check DNS settings and network connectivity. Original error: %s", msg)
	case strings.Contains(msg, "certificate") || strings.Contains(msg, "tls:") || strings.Contains(msg, "x509:"):
		return fmt.Sprintf("TLS/certificate error connecting to WAYSCloud API. Check proxy settings and system clock. Original error: %s", msg)
	case strings.Contains(msg, "connection refused"):
		return fmt.Sprintf("Connection refused by WAYSCloud API. The service may be temporarily unavailable. Original error: %s", msg)
	case strings.Contains(msg, "connection reset"):
		return fmt.Sprintf("Connection reset by WAYSCloud API. Try again. Original error: %s", msg)
	default:
		return fmt.Sprintf("An unexpected error occurred: %s", msg)
	}
}
