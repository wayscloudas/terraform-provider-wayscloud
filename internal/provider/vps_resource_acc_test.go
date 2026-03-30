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

func TestAccVPSResource_tags_labels(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with tags and labels
			{
				Config: `
resource "wayscloud_vps" "tagged" {
  hostname    = "tfacc-tagged.example.com"
  plan_code   = "NO-Start-Linux-2cpu-4096mb-30gb"
  region      = "NO"
  os_template = "ubuntu-24.04"

  tags = ["web", "prod"]

  labels = {
    env  = "prod"
    role = "frontend"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wayscloud_vps.tagged", "tags.#", "2"),
					resource.TestCheckResourceAttr("wayscloud_vps.tagged", "tags.0", "web"),
					resource.TestCheckResourceAttr("wayscloud_vps.tagged", "tags.1", "prod"),
					resource.TestCheckResourceAttr("wayscloud_vps.tagged", "labels.env", "prod"),
					resource.TestCheckResourceAttr("wayscloud_vps.tagged", "labels.role", "frontend"),
				),
			},
			// Step 2: Update tags and labels in-place
			{
				Config: `
resource "wayscloud_vps" "tagged" {
  hostname    = "tfacc-tagged.example.com"
  plan_code   = "NO-Start-Linux-2cpu-4096mb-30gb"
  region      = "NO"
  os_template = "ubuntu-24.04"

  tags = ["web", "staging"]

  labels = {
    env = "staging"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wayscloud_vps.tagged", "tags.#", "2"),
					resource.TestCheckResourceAttr("wayscloud_vps.tagged", "tags.0", "web"),
					resource.TestCheckResourceAttr("wayscloud_vps.tagged", "tags.1", "staging"),
					resource.TestCheckResourceAttr("wayscloud_vps.tagged", "labels.env", "staging"),
					resource.TestCheckNoResourceAttr("wayscloud_vps.tagged", "labels.role"),
				),
			},
			// Step 3: Import state
			{
				ResourceName:      "wayscloud_vps.tagged",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"ssh_keys"},
			},
		},
	})
}
