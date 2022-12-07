package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccExampleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testNeonEndpointResourceConfig1(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("neon_endpoint.test", "branch_id", "branch"),
				),
			},
			// ImportState testing
			/*{
				ResourceName:      "scaffolding_example.test",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"configurable_attribute"},
			},*/
			// Update and Read testing
			/*{
				Config: testNeonEndpointResourceConfig1(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_example.test", "configurable_attribute", "two"),
				),
			},*/
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testNeonEndpointResourceConfig2() string {
	return fmt.Sprintf(`
resource "neon_endpoint" "test" {
	branch_id = "branch"
	type = "read_write"
}
`)
}
func testNeonEndpointResourceConfig1() string {
	return `
provider "neon" {}
resource "neon_endpoint" "test" {
	branch_id = "branch"
	type = "read_write"
	project_id = "project"
	settings = {
		description = "first_description"
		pg_settings = [{
			description = "second_description"
			test_setting = "test_first_setting"
		}, {
			description = "third_description"
			test_setting = "test_second_setting"
		}]
	}
	pooler_enabled = false
	pooler_mode = "transaction"
	disabled = false
	passwordless_access = true
}
`
}
