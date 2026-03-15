// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIoTDeviceResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "wayscloud_iot_device" "test" {
  device_id = "tfacc-sensor-01"
  name      = "Test Sensor"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wayscloud_iot_device.test", "device_id", "tfacc-sensor-01"),
					resource.TestCheckResourceAttr("wayscloud_iot_device.test", "name", "Test Sensor"),
				),
			},
		},
	})
}
