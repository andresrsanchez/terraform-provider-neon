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
				Config: testNeonProjectResourceBasic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("neon_project.test", "name", "name"),
					resource.TestCheckResourceAttr("neon_project.test", "provisioner", "k8s-pod"),
					resource.TestCheckResourceAttr("neon_project.test", "region_id", "aws-us-east-2"),
					resource.TestCheckResourceAttr("neon_project.test", "pg_version", "14"),
					//resource.TestCheckResourceAttr("neon_project.test", "settings", "name"),
				),
			},
			// ImportState testing
			/*{
				ResourceName:      "scaffolding_example.test",
				ImportState:       true,
				ImportStateVerify: true,
				// This is not normally necessary, but is here because this
				// example code does not have an actual upstream service.
				// Once the Read method is able to refresh information from
				// the upstream service, this can be removed.
				ImportStateVerifyIgnore: []string{"configurable_attribute"},
			},*/
			// Update and Read testing
			/*{
				Config: testNeonEndpointResourceConfig1(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("scaffolding_example.test", "configurable_attribute", "two"),
				),
			},*/
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testNeonProjectResourceBasic() string {
	return `
resource "neon_project" "test" {
	name = "name"
	provisioner = "k8s-pod"
	region_id = "aws-us-east-2"
	pg_version = 14	
}
`
}

//	Provisioner              string                `json:"provisioner,omitempty"`
//	RegionID                 string                `json:"region_id,omitempty"`
//	PgVersion                int64                 `json:"pg_version,omitempty"`
//	Settings                 *endpointSettingsJSON `json:"settings,omitempty"`
//	Autoscaling_limit_min_cu int64                 `json:"autoscaling_limit_min_cu,omitempty"`
//	Autoscaling_limit_max_cu int64                 `json:"autoscaling_limit_max_cu,omitempty"`

/*func testNeonProjectResourceComplex_1() string {
	return `
provider "neon" {}
resource "neon_endpoint" "test" {
	branch_id = "branch"
	type = "read_write"
	project_id = "project"
	settings = {
		description = "first_description"
		pg_settings = [{
			description = "second_description"
			test_setting = "test_first_setting"
		}, {
			description = "third_description"
			test_setting = "test_second_setting"
		}]
	}
	pooler_enabled = false
	pooler_mode = "transaction"
	disabled = false
	passwordless_access = true
}
`
}*/
