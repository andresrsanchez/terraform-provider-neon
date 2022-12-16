package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDatabaseResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testNeonDatabaseResourceBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("neon_project.test", "name", "name_project"),
					resource.TestCheckResourceAttr("neon_branch.test", "name", "name_branch"),
					resource.TestCheckResourceAttr("neon_database.test", "name", "name_database"),
				),
			},
		},
	})
}

func testNeonDatabaseResourceBasic() string {
	return `
resource "neon_project" "test" {
	name = "name_project"
}

resource "neon_branch" "test" {
	project_id = neon_project.test.id
	name = "name_branch"
}

resource "neon_database" "test" {
	project_id = neon_project.test.id
	branch_id = neon_branch.test.id
	name = "name_database"
	owner_name = "andresrsanchez"
}
`
}
