package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestEndpointDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testEndpointDataSource(),
				Check:  resource.ComposeAggregateTestCheckFunc(
				//resource.TestCheckResourceAttr("neon_endpoint.test", "branch_id", "branch"),
				),
			},
		},
	})
}

func testEndpointDataSource() string {
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
