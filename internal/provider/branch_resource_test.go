package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestBranchResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testNeonBranchResourceBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("neon_project.test", "name", "name"),
					resource.TestCheckResourceAttr("neon_project.test", "provisioner", "k8s-pod"),
					resource.TestCheckResourceAttr("neon_project.test", "region_id", "aws-us-east-2"),
					resource.TestCheckResourceAttr("neon_project.test", "pg_version", "14"),
				),
			},
		},
	})
}

func testNeonBranchResourceBasic() string {
	return fmt.Sprintf(`
resource "neon_project" "test" {
	name = "name_project"
}

resource "neon_branch" "test" {
	project_id = neon_project.test.id
	name = "name_branch"
}
`)
}
