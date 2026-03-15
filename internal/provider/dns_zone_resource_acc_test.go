// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSZoneResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "wayscloud_dns_zone" "test" {
  name = "tfacc-test.example.com"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wayscloud_dns_zone.test", "name", "tfacc-test.example.com"),
					resource.TestCheckResourceAttrSet("wayscloud_dns_zone.test", "status"),
				),
			},
			{
				ResourceName:      "wayscloud_dns_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
