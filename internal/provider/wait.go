// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// WaitConfig configures polling behavior for async resources
type WaitConfig struct {
	Timeout      time.Duration
	PollInterval time.Duration
	// PendingStates are states that indicate the operation is still in progress
	PendingStates []string
	// TargetStates are states that indicate success
	TargetStates []string
}

// DefaultWaitConfig returns a sensible default for most resources
func DefaultWaitConfig() WaitConfig {
	return WaitConfig{
		Timeout:       10 * time.Minute,
		PollInterval:  15 * time.Second,
		PendingStates: []string{"provisioning", "creating", "building", "deploying"},
		TargetStates:  []string{"active", "running"},
	}
}

// PollFunc is called on each poll iteration.
// Returns the current status string, the response data, and any error.
type PollFunc func(ctx context.Context) (status string, err error)

// WaitForState polls until a target state is reached, with retry tolerance
// for transient poll failures (429, 502, 503, 504).
func WaitForState(ctx context.Context, config WaitConfig, pollFn PollFunc) error {
	deadline := time.Now().Add(config.Timeout)
	consecutiveErrors := 0
	maxConsecutiveErrors := 2

	targetSet := make(map[string]bool, len(config.TargetStates))
	for _, s := range config.TargetStates {
		targetSet[s] = true
	}
	pendingSet := make(map[string]bool, len(config.PendingStates))
	for _, s := range config.PendingStates {
		pendingSet[s] = true
	}

	for time.Now().Before(deadline) {
		status, err := pollFn(ctx)
		if err != nil {
			// Check for 404 — resource was deleted during polling
			var apiErr *client.APIError
			if errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
				return fmt.Errorf("resource was deleted during polling (404)")
			}

			// Check for transient errors — tolerate up to maxConsecutiveErrors
			if errors.As(err, &apiErr) && isTransient(apiErr.StatusCode) {
				consecutiveErrors++
				tflog.Warn(ctx, "Transient poll error, retrying", map[string]interface{}{
					"status_code":        apiErr.StatusCode,
					"consecutive_errors": consecutiveErrors,
				})
				if consecutiveErrors > maxConsecutiveErrors {
					return fmt.Errorf("too many consecutive poll failures: %w", err)
				}
				// Wait and retry
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(config.PollInterval):
					continue
				}
			}

			// Non-transient error
			return err
		}

		// Reset consecutive error counter on success
		consecutiveErrors = 0

		tflog.Debug(ctx, "Poll status", map[string]interface{}{
			"status": status,
		})

		// Check if we've reached a target state
		if targetSet[status] {
			return nil
		}

		// Check for terminal failure states
		if !pendingSet[status] && !targetSet[status] {
			return fmt.Errorf("resource entered unexpected state: %s", status)
		}

		// Wait before next poll
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(config.PollInterval):
			// Continue
		}
	}

	return fmt.Errorf("timeout waiting for resource (waited %s)", config.Timeout)
}

// isTransient returns true for status codes that indicate a transient error
func isTransient(statusCode int) bool {
	return statusCode == 429 || statusCode == 502 || statusCode == 503 || statusCode == 504
}
