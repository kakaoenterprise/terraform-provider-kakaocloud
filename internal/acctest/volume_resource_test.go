// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccVolumeResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + `
resource "kakaocloud_volume" "test" {
  name = "ted-test-by-terraform"
  size = 1
  availability_zone = "kr-central-2-a"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kakaocloud_volume.test", "name", "ted-test-by-terraform"),
					resource.TestCheckResourceAttr("kakaocloud_volume.test", "size", "1"),
					resource.TestCheckResourceAttr("kakaocloud_volume.test", "availability_zone", "kr-central-2-a"),
					// Verify dynamic values have any value set in the state.
					resource.TestCheckResourceAttrSet("kakaocloud_volume.test", "id"),
					resource.TestCheckResourceAttrSet("kakaocloud_volume.test", "status"),
					resource.TestCheckResourceAttrSet("kakaocloud_volume.test", "created_at"),
				),
			},
			{
				ResourceName:      "kakaocloud_volume.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{"attach_status", "created_at", "encryption_key_id", "instance_id",
					"instance_name", "is_root", "launched_at", "mount_point", "previous_status", "project_id", "status",
					"type", "updated_at", "volume_type"},
			},
			// Update and Read testing
			{
				Config: providerConfig + `
resource "kakaocloud_volume" "test" {
  name = "ted-test-x-by-terraform"
  description = "This is a acc test"
  size = 1
  availability_zone = "kr-central-2-a"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("kakaocloud_volume.test", "name", "ted-test-x-by-terraform"),
					resource.TestCheckResourceAttr("kakaocloud_volume.test", "description", "This is a acc test"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}
