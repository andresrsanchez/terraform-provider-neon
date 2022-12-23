package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestBranchDataResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testBranchDataResource(),
				Check:  resource.ComposeAggregateTestCheckFunc(
				//resource.TestCheckResourceAttr("neon_project.test", "name", "name_project"),
				),
			},
		},
	})
}

func testBranchDataResource() string {
	return `
resource "neon_project" "test" {
	name = "name_project"
}

resource "neon_branch" "test" {
	project_id = neon_project.test.id
	name = "name_branch"
	endpoints = [
		{
			type = "read_write"
			autoscaling_limit_min_cu = 1
			autoscaling_limit_max_cu = 1
		}
	]
}

data "neon_branch" "test" {
    project_id = neon_project.test.id
	name = "name_branch"
}
`
}
