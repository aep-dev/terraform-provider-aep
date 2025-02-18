// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccExampleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testExamplePublisherConfig("my-pub", "pub-description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_publisher.my-pub", "description", "pub-description"),
					resource.TestCheckResourceAttr("scaffolding_publisher.my-pub", "path", "/publishers/1"),
				),
			},
			// ImportState testing
			{
				ResourceName:                         "scaffolding_publisher.my-pub",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "path",
			},
			// Update and Read testing
			{
				Config: testExamplePublisherConfig("my-pub", "pub-description2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_publisher.my-pub", "description", "pub-description2"),
					resource.TestCheckResourceAttr("scaffolding_publisher.my-pub", "path", "/publishers/1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testExamplePublisherConfig(publisher string, description string) string {
	f := fmt.Sprintf(`
resource "scaffolding_publisher" %[1]q {
  description = %[2]q
}
`, publisher, description)
	return f
}
