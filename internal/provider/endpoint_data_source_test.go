package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestEndpointResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				//Config: testEndpointResourceCreate(),
				Check: resource.ComposeAggregateTestCheckFunc(
				//resource.TestCheckResourceAttr("data.neon_endpoint.test", "region_id", "aws-us-east-2"),
				),
			},
		},
	})
}

//(neon_project.test): https://github.com/hashicorp/terraform/issues/31527
//cannot create multiple read_write endpoints in a project (for now)
func testEndpointResourceCreate() string {
	return `
resource "neon_project" "test" {
	name = "name_project"
}

resource "neon_endpoint" "test" {
	project_id = neon_project.test.id
	branch_id = (neon_project.test).branch.id
	type = "read_write"
	region_id = "aws-us-east-2"
}

data "neon_endpoint" "test" {
    project_id = neon_project.test.id
	id = neon_endpoint.test.id
}
`
}
