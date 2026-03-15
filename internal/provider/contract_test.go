// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestProvider_contract verifies the provider can be instantiated and passes
// basic schema validation without any API calls.
func TestProvider_contract(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"wayscloud": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				// Empty config — just verify the provider initializes
				Config: `provider "wayscloud" { api_key = "wayscloud_api_test1234_fakesecretkey" }`,
			},
		},
	})
}
