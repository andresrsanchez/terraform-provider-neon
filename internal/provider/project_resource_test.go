package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				//Destroy: true,
				Config: testProjectCreateResource(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("neon_project.test", "name", "name"),
					resource.TestCheckResourceAttr("neon_project.test", "engine", "k8s-pod"),
					resource.TestCheckResourceAttr("neon_project.test", "region_id", "aws-us-east-2"),
					resource.TestCheckResourceAttr("neon_project.test", "pg_version", "14"),
					//resource.TestCheckResourceAttr("neon_project.test", "settings", "name"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "neon_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testProjectUpdateResource(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("neon_project.test", "name", "name_updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testProjectCreateResource() string {
	return `
resource "neon_project" "test" {
	name = "name"
	engine = "k8s-pod"
	region_id = "aws-us-east-2"
	pg_version = 14	
}
`
}

func testProjectUpdateResource() string {
	return `
resource "neon_project" "test" {
	name = "name_updated"
}
`
}
