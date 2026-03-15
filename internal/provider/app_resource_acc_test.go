// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAppResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "wayscloud_app" "test" {
  name   = "tfacc-app-test"
  region = "no"
  plan   = "app-basic"

  env_vars = {
    NODE_ENV = "test"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wayscloud_app.test", "name", "tfacc-app-test"),
					resource.TestCheckResourceAttrSet("wayscloud_app.test", "id"),
					resource.TestCheckResourceAttr("wayscloud_app.test", "region", "no"),
					resource.TestCheckResourceAttr("wayscloud_app.test", "env_vars.NODE_ENV", "test"),
				),
			},
			{
				ResourceName:            "wayscloud_app.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"env_vars"},
			},
		},
	})
}
