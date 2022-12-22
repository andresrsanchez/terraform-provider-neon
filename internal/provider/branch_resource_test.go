package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestBranchResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testBranchResourceCreate(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("neon_project.test", "name", "name_project"),
					resource.TestCheckResourceAttr("neon_branch.test", "name", "name_branch"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "neon_branch.test",
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					if rs, ok := s.RootModule().Resources["neon_branch.test"]; ok {
						return fmt.Sprintf("%s/%s", rs.Primary.ID, rs.Primary.Attributes["project_id"]), nil
					}
					return "", fmt.Errorf("cannot find neon_branch.test")
				},
			},
			// Update and Read testing
			{
				Config: testBranchResourceUpdate(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("neon_branch.test", "name", "name_branch_updated"),
				),
			},
		},
	})
}

func testBranchResourceCreate() string {
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
`
}

func testBranchResourceUpdate() string {
	return `
resource "neon_project" "test" {
	name = "name_project"
}

resource "neon_branch" "test" {
	project_id = neon_project.test.id
	name = "name_branch_updated"
	endpoints = [
		{
			type = "read_write"
			autoscaling_limit_min_cu = 1
			autoscaling_limit_max_cu = 1
		}
	]
}
`
}
