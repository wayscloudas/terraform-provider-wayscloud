// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSMSSenderProfileResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "wayscloud_sms_sender_profile" "test" {
  name        = "tfacc-sender"
  sender_name = "TFAccTest"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wayscloud_sms_sender_profile.test", "name", "tfacc-sender"),
					resource.TestCheckResourceAttrSet("wayscloud_sms_sender_profile.test", "id"),
				),
			},
		},
	})
}
