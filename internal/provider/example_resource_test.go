// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPublisherResource(t *testing.T) {
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

func TestBookResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testExampleBookConfig("my-book", "1", "my-pub"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_publisher.my-pub", "description", "pub-description"),
					resource.TestCheckResourceAttr("scaffolding_publisher.my-pub", "path", "/publishers/1"),
					resource.TestCheckResourceAttr("scaffolding_book.my-book", "price", "1"),
					resource.TestCheckResourceAttr("scaffolding_book.my-book", "path", "/publishers/1/books/1"),
				),
			},
			// Update and Read testing
			{
				Config: testExampleBookConfig("my-book", "3", "my-pub"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_publisher.my-pub", "description", "pub-description"),
					resource.TestCheckResourceAttr("scaffolding_publisher.my-pub", "path", "/publishers/1"),
					resource.TestCheckResourceAttr("scaffolding_book.my-book", "price", "3"),
					resource.TestCheckResourceAttr("scaffolding_book.my-book", "path", "/publishers/1/books/1"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testExamplePublisherConfig(publisher string, description string) string {
	return fmt.Sprintf(`
resource "scaffolding_publisher" %[1]q {
  description = %[2]q
}
`, publisher, description)
}

func testExampleBookConfig(book string, price string, publisher string) string {
	return fmt.Sprintf(`
resource "scaffolding_publisher" %[3]q {
  description = "pub-description"
}

resource "scaffolding_book" %[1]q {
  price = %[2]q
  publisher = scaffolding_publisher.%[3]s.path
  published = true
  isbn = ["1234"]
}
`, book, price, publisher)
}
