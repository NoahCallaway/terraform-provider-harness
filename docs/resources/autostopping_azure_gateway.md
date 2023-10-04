---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "harness_autostopping_azure_gateway Resource - terraform-provider-harness"
subcategory: "Next Gen"
description: |-
  Resource for creating an Azure Application Gateway
---

# harness_autostopping_azure_gateway (Resource)

Resource for creating an Azure Application Gateway

## Example Usage

```terraform
resource "harness_autostopping_azure_gateway" "test" {
  name               = "name"
  cloud_connector_id = "cloud_connector_id"
  host_name          = "host_name"
  region             = "eastus2"
  resource_group     = "resource_group"
  subnet_id          = "/subscriptions/subscription_id/resourceGroups/resource_group/providers/Microsoft.Network/virtualNetworks/virtual_network/subnets/subnet_id"
  vpc                = "/subscriptions/subscription_id/resourceGroups/resource_group/providers/Microsoft.Network/virtualNetworks/virtual_network"
  azure_func_region  = "westus2"
  frontend_ip        = "/subscriptions/subscription_id/resourceGroups/resource_group/providers/Microsoft.Network/publicIPAddresses/publicip"
  sku_size           = "sku2"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `azure_func_region` (String) Region in which azure cloud function will be provisioned
- `cloud_connector_id` (String) Id of the cloud connector
- `frontend_ip` (String)
- `host_name` (String) Hostname for the proxy
- `name` (String) Name of the proxy
- `region` (String) Region in which cloud resources are hosted
- `resource_group` (String) Resource group in which cloud resources are hosted
- `sku_size` (String) Size of machine used for the gateway
- `subnet_id` (String) Subnet in which cloud resources are hosted
- `vpc` (String) VPC in which cloud resources are hosted

### Read-Only

- `id` (String) The ID of this resource.
- `identifier` (String) Unique identifier of the resource