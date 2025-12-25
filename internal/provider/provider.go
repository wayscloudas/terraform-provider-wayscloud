// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// Ensure WAYSCloudProvider satisfies various provider interfaces.
var _ provider.Provider = &WAYSCloudProvider{}

// WAYSCloudProvider defines the provider implementation.
type WAYSCloudProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// WAYSCloudProviderModel describes the provider data model.
type WAYSCloudProviderModel struct {
	APIKey   types.String `tfsdk:"api_key"`
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *WAYSCloudProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "wayscloud"
	resp.Version = p.version
}

func (p *WAYSCloudProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
The WAYSCloud provider allows you to manage WAYSCloud resources using Terraform.

## Authentication

The provider supports authentication via API Key. You can configure it in three ways:

1. **Provider configuration** (not recommended for production):
` + "```hcl" + `
provider "wayscloud" {
  api_key = "wayscloud_api_xxx..."
}
` + "```" + `

2. **Environment variable** (recommended):
` + "```bash" + `
export WAYSCLOUD_API_KEY="wayscloud_api_xxx..."
` + "```" + `

3. **Terraform variable**:
` + "```hcl" + `
variable "wayscloud_api_key" {
  type      = string
  sensitive = true
}

provider "wayscloud" {
  api_key = var.wayscloud_api_key
}
` + "```" + `

## Getting an API Key

1. Log in to [my.wayscloud.services](https://my.wayscloud.services)
2. Navigate to **Account** → **API Keys**
3. Click **Create API Key**
4. Select the services you need access to (DNS, VPS, Storage, etc.)
5. Copy the generated key immediately (it's only shown once)

## Example Usage

` + "```hcl" + `
terraform {
  required_providers {
    wayscloud = {
      source  = "wayscloud/wayscloud"
      version = "~> 0.1.0"
    }
  }
}

provider "wayscloud" {}

resource "wayscloud_dns_zone" "example" {
  name = "example.com"
}

resource "wayscloud_dns_record" "www" {
  zone_name = wayscloud_dns_zone.example.name
  name      = "www"
  type      = "A"
  value     = "192.0.2.1"
  ttl       = 300
}
` + "```" + `
`,
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "WAYSCloud API Key for authentication. Can also be set via `WAYSCLOUD_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "WAYSCloud API endpoint. Defaults to `https://api.wayscloud.services`. Can also be set via `WAYSCLOUD_ENDPOINT` environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *WAYSCloudProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring WAYSCloud provider")

	var config WAYSCloudProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check environment variables for defaults
	apiKey := os.Getenv("WAYSCLOUD_API_KEY")
	endpoint := os.Getenv("WAYSCLOUD_ENDPOINT")

	// Override with explicit configuration if provided
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}

	// Validate required configuration
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The provider cannot create the WAYSCloud client because the API key is missing. "+
				"Set it in the provider configuration or via the WAYSCLOUD_API_KEY environment variable.",
		)
		return
	}

	// Create WAYSCloud client
	c := client.NewClient(apiKey, endpoint)

	tflog.Debug(ctx, "Created WAYSCloud client", map[string]interface{}{
		"endpoint": c.BaseURL,
	})

	// Make the client available to resources and data sources
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *WAYSCloudProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDNSZoneResource,
		NewDNSRecordResource,
		NewRedisInstanceResource,
		NewS3BucketResource,
		NewDatabaseResource,
		NewVPSResource,
		NewAppResource,
	}
}

func (p *WAYSCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Data sources will be added here
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &WAYSCloudProvider{
			version: version,
		}
	}
}
