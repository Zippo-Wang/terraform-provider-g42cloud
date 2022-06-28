package modelarts

import (
	"fmt"
	"testing"

	"github.com/g42cloud-terraform/terraform-provider-g42cloud/g42cloud/services/acceptance"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccNotebookImages_basic(t *testing.T) {
	dataSourceName := "data.g42cloud_modelarts_notebook_images.test"

	dc := acceptance.InitDataSourceCheck(dataSourceName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceImages_basic("BUILD_IN", "x86_64"),
				Check: resource.ComposeTestCheckFunc(
					dc.CheckResourceExists(),
					resource.TestCheckResourceAttr(dataSourceName, "images.0.type", "BUILD_IN"),
					resource.TestCheckResourceAttr(dataSourceName, "images.0.cpu_arch", "x86_64"),
					resource.TestCheckResourceAttrSet(dataSourceName, "images.0.id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "images.0.name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "images.0.swr_path"),
					resource.TestCheckResourceAttrSet(dataSourceName, "images.0.description"),
				),
			},
		},
	})
}

func testAccDataSourceImages_basic(imageType, cpuArch string) string {
	return fmt.Sprintf(`
data "g42cloud_modelarts_notebook_images" "test" {
  type     = "%s"
  cpu_arch = "%s"
}
`, imageType, cpuArch)
}
