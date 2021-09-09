package provider

import (
	"fmt"
	"testing"

	"github.com/harness-io/harness-go-sdk/harness/api"
	"github.com/harness-io/harness-go-sdk/harness/api/cac"
	"github.com/harness-io/harness-go-sdk/harness/helpers"
	"github.com/harness-io/harness-go-sdk/harness/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func TestAccResourceAzureCloudProviderConnector(t *testing.T) {

	var (
		name         = fmt.Sprintf("%s_%s", t.Name(), utils.RandStringBytes(4))
		updatedName  = fmt.Sprintf("%s_updated", name)
		resourceName = "harness_cloudprovider_azure.test"
	)

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		CheckDestroy:      testAccCloudProviderDestroy(resourceName),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAzureCloudProvider(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					testAccCheckAzureCloudProviderExists(t, resourceName, name),
				),
			},
			{
				Config: testAccResourceAzureCloudProvider(updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAzureCloudProviderExists(t, resourceName, name),
				),
			},
		},
	})
}

func TestAccResourceAzureCloudProviderConnector_DeleteUnderlyingResource(t *testing.T) {

	var (
		name         = fmt.Sprintf("%s_%s", t.Name(), utils.RandStringBytes(4))
		resourceName = "harness_cloudprovider_azure.test"
	)

	resource.UnitTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceAzureCloudProvider(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					testAccCheckAzureCloudProviderExists(t, resourceName, name),
				),
			},
			{
				PreConfig: func() {
					testAccConfigureProvider()
					c := testAccProvider.Meta().(*api.Client)
					cp, err := c.CloudProviders().GetAzureCloudProviderByName(name)
					require.NoError(t, err)
					require.NotNil(t, cp)

					err = c.CloudProviders().DeleteCloudProvider(cp.Id)
					require.NoError(t, err)
				},
				Config:             testAccResourceAzureCloudProvider(name),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccResourceAzureCloudProvider(name string) string {
	return fmt.Sprintf(`
		data "harness_secret_manager" "default" {
			default = true
		}

		resource "harness_encrypted_text" "test" {
			name = "%[1]s"
			value = "%[2]s"
			secret_manager_id = data.harness_secret_manager.default.id
		}

		resource "harness_cloudprovider_azure" "test" {
			name = "%[1]s"
			client_id = "%[3]s"
			tenant_id = "%[4]s"
			key = harness_encrypted_text.test.name
		}
`, name, helpers.TestEnvVars.AzureClientSecret.Get(), helpers.TestEnvVars.AzureClientId.Get(), helpers.TestEnvVars.AzureTenantId.Get())
}

func testAccCheckAzureCloudProviderExists(t *testing.T, resourceName, cloudProviderName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		cp := &cac.AzureCloudProvider{}
		err := testAccGetCloudProvider(resourceName, state, cp)
		if err != nil {
			return err
		}
		return nil
	}
}
