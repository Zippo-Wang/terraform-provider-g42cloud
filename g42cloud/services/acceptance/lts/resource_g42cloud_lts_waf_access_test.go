package lts

import (
	"fmt"
	"strings"
	"testing"

	"github.com/g42cloud-terraform/terraform-provider-g42cloud/g42cloud/services/acceptance/common"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/chnsz/golangsdk"
	"github.com/g42cloud-terraform/terraform-provider-g42cloud/g42cloud/services/acceptance"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
)

func getWAFAccessResourceFunc(cfg *config.Config, state *terraform.ResourceState) (interface{}, error) {
	var (
		region  = acceptance.G42_REGION_NAME
		httpUrl = "v1/{project_id}/waf/config/lts"
		product = "waf"
	)
	client, err := cfg.NewServiceClient(product, region)
	if err != nil {
		return nil, fmt.Errorf("error creating WAF client: %s", err)
	}

	getPath := client.Endpoint + httpUrl
	getPath = strings.ReplaceAll(getPath, "{project_id}", client.ProjectID)
	if epsId := state.Primary.Attributes["enterprise_project_id"]; epsId != "" {
		getPath += fmt.Sprintf("?enterprise_project_id=%s", epsId)
	}

	getOpt := golangsdk.RequestOpts{
		KeepResponseBody: true,
	}

	getResp, err := client.Request("GET", getPath, &getOpt)
	if err != nil {
		return nil, err
	}

	getRespBody, err := utils.FlattenResponse(getResp)
	if err != nil {
		return nil, err
	}

	enabled := utils.PathSearch("enabled", getRespBody, false).(bool)
	if !enabled {
		// the WAF access is not exist
		return nil, golangsdk.ErrDefault404{}
	}

	return getRespBody, nil
}

func TestAccWAFAccess_basic(t *testing.T) {
	var obj interface{}

	name := acceptance.RandomAccResourceName()
	rName := "g42cloud_lts_waf_access.test"

	rc := acceptance.InitResourceCheck(
		rName,
		&obj,
		getWAFAccessResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acceptance.TestAccPreCheck(t) },
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testWAFAccess_basic(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttrPair(rName, "lts_group_id",
						"g42cloud_lts_group.groupA", "id"),
					resource.TestCheckResourceAttrPair(rName, "lts_attack_stream_id",
						"g42cloud_lts_stream.streamA1", "id"),
					resource.TestCheckResourceAttrPair(rName, "lts_access_stream_id",
						"g42cloud_lts_stream.streamA2", "id"),
				),
			},
			{
				Config: testWAFAccess_basic_update1(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttrPair(rName, "lts_group_id",
						"g42cloud_lts_group.groupB", "id"),
					resource.TestCheckResourceAttrPair(rName, "lts_attack_stream_id",
						"g42cloud_lts_stream.streamB1", "id"),
					resource.TestCheckResourceAttrPair(rName, "lts_access_stream_id",
						"g42cloud_lts_stream.streamB2", "id"),
				),
			},
			{
				Config: testWAFAccess_basic_update2(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttrPair(rName, "lts_group_id",
						"g42cloud_lts_group.groupB", "id"),
					resource.TestCheckResourceAttr(rName, "lts_attack_stream_id", ""),
					resource.TestCheckResourceAttr(rName, "lts_access_stream_id", ""),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testWAFAccessImportState(rName),
			},
		},
	})
}

func TestAccWAFAccess_withEpsId(t *testing.T) {
	var obj interface{}

	name := acceptance.RandomAccResourceName()
	rName := "g42cloud_lts_waf_access.test"

	rc := acceptance.InitResourceCheck(
		rName,
		&obj,
		getWAFAccessResourceFunc,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acceptance.TestAccPreCheck(t)
			acceptance.TestAccPreCheckEpsID(t)
		},
		ProviderFactories: acceptance.TestAccProviderFactories,
		CheckDestroy:      rc.CheckResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testWAFAccess_withEpsId(name),
				Check: resource.ComposeTestCheckFunc(
					rc.CheckResourceExists(),
					resource.TestCheckResourceAttr(rName, "enterprise_project_id",
						acceptance.G42_ENTERPRISE_PROJECT_ID_TEST),
					resource.TestCheckResourceAttrPair(rName, "lts_group_id",
						"g42cloud_lts_group.groupA", "id"),
					resource.TestCheckResourceAttrPair(rName, "lts_attack_stream_id",
						"g42cloud_lts_stream.streamA1", "id"),
					resource.TestCheckResourceAttrPair(rName, "lts_access_stream_id",
						"g42cloud_lts_stream.streamA2", "id"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testWAFAccessImportState(rName),
			},
		},
	})
}

func testWAFAccess_base(name string) string {
	return fmt.Sprintf(`
resource "g42cloud_lts_group" "groupA" {
  group_name  = "%[1]s_a"
  ttl_in_days = 30
}

resource "g42cloud_lts_stream" "streamA1" {
  group_id    = g42cloud_lts_group.groupA.id
  stream_name = "%[1]s_a1"
}

resource "g42cloud_lts_stream" "streamA2" {
  group_id    = g42cloud_lts_group.groupA.id
  stream_name = "%[1]s_a2"
}

resource "g42cloud_lts_group" "groupB" {
  group_name  = "%[1]s_b"
  ttl_in_days = 30
}

resource "g42cloud_lts_stream" "streamB1" {
  group_id    = g42cloud_lts_group.groupB.id
  stream_name = "%[1]s_b1"
}

resource "g42cloud_lts_stream" "streamB2" {
  group_id    = g42cloud_lts_group.groupB.id
  stream_name = "%[1]s_b2"
}

%[2]s

resource "g42cloud_waf_dedicated_instance" "test" {
  name               = "%[1]s"
  available_zone     = data.g42cloud_availability_zones.test.names[1]
  specification_code = "waf.instance.professional"
  ecs_flavor         = data.g42cloud_compute_flavors.test.ids[0]
  vpc_id             = g42cloud_vpc.test.id
  subnet_id          = g42cloud_vpc_subnet.test.id
  
  security_group = [
    g42cloud_networking_secgroup.test.id
  ]
}
`, name, common.TestBaseComputeResources(name))
}

func testWAFAccess_epsId(name, epsId string) string {
	return fmt.Sprintf(`
resource "g42cloud_lts_group" "groupA" {
  group_name  = "%[1]s_a"
  ttl_in_days = 30
}

resource "g42cloud_lts_stream" "streamA1" {
  group_id    = g42cloud_lts_group.groupA.id
  stream_name = "%[1]s_a1"
}

resource "g42cloud_lts_stream" "streamA2" {
  group_id    = g42cloud_lts_group.groupA.id
  stream_name = "%[1]s_a2"
}

%[2]s

resource "g42cloud_waf_dedicated_instance" "test" {
  name                  = "%[1]s"
  available_zone        = data.g42cloud_availability_zones.test.names[1]
  specification_code    = "waf.instance.professional"
  ecs_flavor            = data.g42cloud_compute_flavors.test.ids[0]
  enterprise_project_id = "%[3]s"
  vpc_id                = g42cloud_vpc.test.id
  subnet_id             = g42cloud_vpc_subnet.test.id
  
  security_group = [
    g42cloud_networking_secgroup.test.id
  ]
}
`, name, common.TestBaseComputeResources(name), epsId)
}

func testWAFAccess_basic(name string) string {
	return fmt.Sprintf(`
%[1]s

resource "g42cloud_lts_waf_access" "test" {
  lts_group_id         = g42cloud_lts_group.groupA.id
  lts_attack_stream_id = g42cloud_lts_stream.streamA1.id
  lts_access_stream_id = g42cloud_lts_stream.streamA2.id

  depends_on = [
    g42cloud_waf_dedicated_instance.test
  ]
}
`, testWAFAccess_base(name))
}

func testWAFAccess_basic_update1(name string) string {
	return fmt.Sprintf(`
%[1]s

resource "g42cloud_lts_waf_access" "test" {
  lts_group_id         = g42cloud_lts_group.groupB.id
  lts_attack_stream_id = g42cloud_lts_stream.streamB1.id
  lts_access_stream_id = g42cloud_lts_stream.streamB2.id
}
`, testWAFAccess_base(name))
}

func testWAFAccess_basic_update2(name string) string {
	return fmt.Sprintf(`
%[1]s

resource "g42cloud_lts_waf_access" "test" {
  lts_group_id = g42cloud_lts_group.groupB.id
}
`, testWAFAccess_base(name))
}

func testWAFAccess_withEpsId(name string) string {
	return fmt.Sprintf(`
%[1]s

resource "g42cloud_lts_waf_access" "test" {
  enterprise_project_id = "%[2]s"
  lts_group_id          = g42cloud_lts_group.groupA.id
  lts_attack_stream_id  = g42cloud_lts_stream.streamA1.id
  lts_access_stream_id  = g42cloud_lts_stream.streamA2.id

  depends_on = [
    g42cloud_waf_dedicated_instance.test
  ]
}
`, testWAFAccess_epsId(name, acceptance.G42_ENTERPRISE_PROJECT_ID_TEST), acceptance.G42_ENTERPRISE_PROJECT_ID_TEST)
}

func testWAFAccessImportState(name string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return "", fmt.Errorf("resource (%s) not found: %s", name, rs)
		}

		epsId := rs.Primary.Attributes["enterprise_project_id"]
		if epsId == "" {
			// the default enterprise project ID is `0`
			epsId = "0"
		}
		return epsId, nil
	}
}
