// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories returns the provider factories for acceptance testing
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"wayscloud": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates that required environment variables are set
func testAccPreCheck(t *testing.T) {
	t.Helper()

	if v := os.Getenv("WAYSCLOUD_API_KEY"); v == "" {
		t.Fatal("WAYSCLOUD_API_KEY must be set for acceptance tests")
	}
}
