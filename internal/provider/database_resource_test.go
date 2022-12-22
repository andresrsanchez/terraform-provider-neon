package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestDatabaseResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testDatabaseCreateResource(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("neon_project.test", "name", "name_project"),
					resource.TestCheckResourceAttr("neon_branch.test", "name", "name_branch"),
					resource.TestCheckResourceAttr("neon_database.test", "name", "name_database"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "neon_database.test",
				ImportState:       true,
				ImportStateVerify: false,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					if rs, ok := s.RootModule().Resources["neon_database.test"]; ok {
						return fmt.Sprintf("%s/%s/%s", rs.Primary.Attributes["project_id"], rs.Primary.Attributes["branch_id"], rs.Primary.Attributes["name"]), nil
					}
					return "", fmt.Errorf("cannot find neon_database.test")
				},
			},
			// Update and Read testing
			{
				Config: testDatabaseUpdateResource(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("neon_database.test", "name", "name_database_updated"),
				),
			},
		},
	})
}

func testDatabaseCreateResource() string {
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

resource "neon_database" "test" {
	project_id = neon_project.test.id
	branch_id = neon_branch.test.id
	name = "name_database"
	owner_name = "andresrsanchez"
	
}
`
}

func testDatabaseUpdateResource() string {
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

resource "neon_database" "test" {
	project_id = neon_project.test.id
	branch_id = neon_branch.test.id
	name = "name_database_updated"
	owner_name = "andresrsanchez"
	
}
`
}
