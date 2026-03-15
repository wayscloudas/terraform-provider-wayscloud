// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSRecordResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "wayscloud_dns_zone" "test" {
  name = "tfacc-record.example.com"
}

resource "wayscloud_dns_record" "test" {
  zone_name = wayscloud_dns_zone.test.name
  name      = "www"
  type      = "A"
  value     = "192.0.2.1"
  ttl       = 300
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wayscloud_dns_record.test", "name", "www"),
					resource.TestCheckResourceAttr("wayscloud_dns_record.test", "type", "A"),
					resource.TestCheckResourceAttr("wayscloud_dns_record.test", "value", "192.0.2.1"),
					resource.TestCheckResourceAttr("wayscloud_dns_record.test", "ttl", "300"),
				),
			},
		},
	})
}
