// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVPSResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "wayscloud_vps" "test" {
  hostname    = "tfacc-test.example.com"
  plan_code   = "NO-Start-Linux-2cpu-4096mb-30gb"
  region      = "NO"
  os_template = "ubuntu-24.04"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wayscloud_vps.test", "hostname", "tfacc-test.example.com"),
					resource.TestCheckResourceAttrSet("wayscloud_vps.test", "id"),
					resource.TestCheckResourceAttrSet("wayscloud_vps.test", "status"),
				),
			},
			{
				ResourceName:      "wayscloud_vps.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"ssh_keys"},
			},
		},
	})
}
