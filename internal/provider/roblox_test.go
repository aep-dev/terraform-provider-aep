// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Debugging the Terraform provider is really hard!
// This test exists so that I can get a debugger with the Roblox API.
// This isn't meant to be a real test.
func TestRobloxDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		PreCheck:                 func() {},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithRoblox,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testDSData(),
			},
		},
	})
}

func testDSData() string {
	return fmt.Sprintf(`
provider "scaffolding" {
  headers = {}
}

data "scaffolding_data-store-entry" "ds" {
  universe_id = "null"
  data_store_id = "loop-test"	
}
`)
}
