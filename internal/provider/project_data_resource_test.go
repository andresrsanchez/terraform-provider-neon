package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestProjectDataResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				//Destroy: true,
				Config: testProjectDataSource(),
				Check:  resource.ComposeAggregateTestCheckFunc(
				//resource.TestCheckResourceAttr("neon_project.test", "settings", "name"),
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
`
}
