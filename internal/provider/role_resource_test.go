package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestRoleResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testRoleResourceCreate(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("neon_project.test", "name", "name_project"),
					resource.TestCheckResourceAttr("neon_branch.test", "name", "name_branch"),
					resource.TestCheckResourceAttr("neon_role.test", "name", "name_role"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "neon_role.test",
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					if rs, ok := s.RootModule().Resources["neon_role.test"]; ok {
						return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["project_id"], rs.Primary.Attributes["branch_id"], rs.Primary.Attributes["name"]), nil
					}
					return "", fmt.Errorf("cannot find neon_role.test")
				},
			},
		},
	})
}

func testRoleResourceCreate() string {
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

resource "neon_role" "test" {
	project_id = neon_project.test.id
	branch_id = neon_branch.test.id
	name = "name_role"
}
`
}
