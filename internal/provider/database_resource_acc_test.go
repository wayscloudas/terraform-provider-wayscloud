// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatabaseResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "wayscloud_database" "test" {
  name = "tfacc-db-test"
  type = "postgresql"
  tier = "standard"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wayscloud_database.test", "name", "tfacc-db-test"),
					resource.TestCheckResourceAttr("wayscloud_database.test", "type", "postgresql"),
					resource.TestCheckResourceAttr("wayscloud_database.test", "tier", "standard"),
					resource.TestCheckResourceAttrSet("wayscloud_database.test", "host"),
				),
			},
			{
				ResourceName:            "wayscloud_database.test",
				ImportState:             true,
				ImportStateId:           "postgresql/standard/tfacc-db-test",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "connection_string"},
			},
		},
	})
}
