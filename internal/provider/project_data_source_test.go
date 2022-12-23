package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestProjectDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProjectDataSource(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.neon_project.test", "region_id", "aws-us-east-2"),
				),
			},
		},
	})
}

func testProjectDataSource() string {
	return `
resource "neon_project" "test" {
	name = "name"
	engine = "k8s-pod"
	region_id = "aws-us-east-2"
	pg_version = 14	
}

data "neon_project" "test" {
	id = neon_project.test.id
}
`
}
