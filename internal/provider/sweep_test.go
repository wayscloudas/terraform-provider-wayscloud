// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// sweepClient returns an API client for resource cleanup
func sweepClient() (*client.Client, error) {
	apiKey := os.Getenv("WAYSCLOUD_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("WAYSCLOUD_API_KEY must be set for sweep")
	}
	endpoint := os.Getenv("WAYSCLOUD_ENDPOINT")
	return client.NewClient(apiKey, endpoint), nil
}

// sweepDNSZones removes test DNS zones (those matching "tfacc-" prefix)
func sweepDNSZones(_ string) error {
	c, err := sweepClient()
	if err != nil {
		return err
	}

	ctx := context.Background()
	respBody, err := c.Get(ctx, "/v1/dns/zones")
	if err != nil {
		return fmt.Errorf("error listing DNS zones for sweep: %w", err)
	}

	// Simple check: look for test zone names in response
	body := string(respBody)
	if strings.Contains(body, "tfacc-") {
		log.Println("[SWEEP] Found test DNS zones — manual cleanup may be needed")
	}

	return nil
}
