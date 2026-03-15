// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// TestSweep_DNSZones is a manual test for cleaning up leftover test resources.
// Run with: go test -v -run TestSweep_DNSZones -timeout 5m
func TestSweep_DNSZones(t *testing.T) {
	if os.Getenv("WAYSCLOUD_SWEEP_ENABLED") != "1" {
		t.Skip("WAYSCLOUD_SWEEP_ENABLED must be set to 1")
	}

	c, err := sweepClient()
	if err != nil {
		t.Fatalf("Failed to create sweep client: %s", err)
	}

	if err := sweepDNSZones(c); err != nil {
		t.Fatalf("DNS zone sweep failed: %s", err)
	}
}

func sweepClient() (*client.Client, error) {
	apiKey := os.Getenv("WAYSCLOUD_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("WAYSCLOUD_API_KEY must be set for sweep")
	}
	endpoint := os.Getenv("WAYSCLOUD_ENDPOINT")
	return client.NewClient(apiKey, endpoint), nil
}

func sweepDNSZones(c *client.Client) error {
	ctx := context.Background()
	respBody, err := c.Get(ctx, "/v1/dns/zones")
	if err != nil {
		return fmt.Errorf("error listing DNS zones for sweep: %w", err)
	}

	body := string(respBody)
	if strings.Contains(body, "tf-acc-test-") || strings.Contains(body, "tf-test-") {
		log.Println("[SWEEP] Found test DNS zones — manual cleanup may be needed")
	}

	return nil
}
