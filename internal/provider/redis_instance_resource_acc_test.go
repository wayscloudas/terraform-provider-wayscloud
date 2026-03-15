// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRedisInstanceResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "wayscloud_redis_instance" "test" {
  name   = "tfacc-redis-test"
  region = "no"
  plan   = "redis-starter"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wayscloud_redis_instance.test", "name", "tfacc-redis-test"),
					resource.TestCheckResourceAttrSet("wayscloud_redis_instance.test", "id"),
					resource.TestCheckResourceAttrSet("wayscloud_redis_instance.test", "status"),
				),
			},
			{
				ResourceName:            "wayscloud_redis_instance.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}
