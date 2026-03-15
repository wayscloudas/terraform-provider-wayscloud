// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccS3BucketResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "wayscloud_s3_bucket" "test" {
  bucket_name = "tfacc-bucket-test"
  region      = "no"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wayscloud_s3_bucket.test", "bucket_name", "tfacc-bucket-test"),
					resource.TestCheckResourceAttrSet("wayscloud_s3_bucket.test", "endpoint"),
				),
			},
			{
				ResourceName:            "wayscloud_s3_bucket.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret_key"},
			},
		},
	})
}
