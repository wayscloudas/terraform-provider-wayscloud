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
	PATToken types.String `tfsdk:"pat_token"`
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

The provider supports two authentication methods:

- **API Key** (` + "`api_key`" + ` / ` + "`WAYSCLOUD_API_KEY`" + `): For DNS, VPS, Storage, Redis, IoT, SMS, and App resources.
- **PAT Token** (` + "`pat_token`" + ` / ` + "`WAYSCLOUD_PAT_TOKEN`" + `): For Database and Domain Verification resources.

Both can be configured simultaneously for full access to all resources.

### Environment variables (recommended)

` + "```bash" + `
export WAYSCLOUD_API_KEY="wayscloud_api_xxx..."
export WAYSCLOUD_PAT_TOKEN="wayscloud_pat_xxx..."
` + "```" + `

### Provider configuration

` + "```hcl" + `
provider "wayscloud" {
  api_key   = var.wayscloud_api_key    # DNS, VPS, Storage, Redis, IoT, SMS, Apps
  pat_token = var.wayscloud_pat_token  # Database, Domain Verification
}
` + "```" + `

## Getting Credentials

1. **API Key:** Log in to [my.wayscloud.services](https://my.wayscloud.services) → Account → API Keys
2. **PAT Token:** Log in to [my.wayscloud.services](https://my.wayscloud.services) → Account → Personal Access Tokens

## Example Usage

` + "```hcl" + `
terraform {
  required_providers {
    wayscloud = {
      source  = "wayscloudas/wayscloud"
      version = "~> 0.4"
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
				MarkdownDescription: "WAYSCloud API Key for authentication. Used for DNS, VPS, Storage, Redis, IoT, SMS, and App resources. Can also be set via `WAYSCLOUD_API_KEY` environment variable. If a PAT token (`wayscloud_pat_...`) is provided here and no separate `pat_token` is set, it will be used for both API key and PAT authentication.",
				Optional:            true,
				Sensitive:           true,
			},
			"pat_token": schema.StringAttribute{
				MarkdownDescription: "WAYSCloud Personal Access Token (PAT) for resources that require PAT authentication (Database, Domain Verification). Can also be set via `WAYSCLOUD_PAT_TOKEN` environment variable. Format: `wayscloud_pat_xxx...`.",
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
	patToken := os.Getenv("WAYSCLOUD_PAT_TOKEN")
	endpoint := os.Getenv("WAYSCLOUD_ENDPOINT")

	// Override with explicit configuration if provided
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}
	if !config.PATToken.IsNull() {
		patToken = config.PATToken.ValueString()
	}
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}

	// Validate required configuration
	if apiKey == "" && patToken == "" {
		resp.Diagnostics.AddError(
			"Missing Authentication",
			"The provider requires at least one of: api_key (WAYSCLOUD_API_KEY) or pat_token (WAYSCLOUD_PAT_TOKEN). "+
				"Use api_key for DNS, VPS, Storage, Redis, IoT, SMS, and App resources. "+
				"Use pat_token for Database and Domain Verification resources. "+
				"Both can be configured simultaneously for full access.",
		)
		return
	}

	// Create WAYSCloud client with dual auth support
	c := client.NewClientDualAuth(apiKey, patToken, endpoint)

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
		NewVPSResource,
		NewDatabaseResource,
		NewRedisInstanceResource,
		NewS3BucketResource,
		NewS3BucketKeyResource,
		NewAppResource,
		NewIoTDeviceResource,
		NewIoTDeviceGroupResource,
		NewIoTRuleResource,
		NewDomainVerificationResource,
	}
}

func (p *WAYSCloudProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRegionsDataSource,
		NewDNSZonesDataSource,
		NewVPSPlansDataSource,
		NewVPSOSTemplatesDataSource,
		NewAppPlansDataSource,
		NewDatabaseTypesDataSource,
		NewRedisPlansDataSource,
		NewStorageTiersDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &WAYSCloudProvider{
			version: version,
		}
	}
}
