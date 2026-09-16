// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// regionsPath is the public region list. It needs no authentication.
const regionsPath = "/v1/regions"

// regionStatusActive is the only status in which a region offers services.
const regionStatusActive = "active"

var _ datasource.DataSource = &RegionsDataSource{}

func NewRegionsDataSource() datasource.DataSource {
	return &RegionsDataSource{}
}

type RegionsDataSource struct {
	client *client.Client
}

type RegionsDataSourceModel struct {
	Regions []RegionModel `tfsdk:"regions"`
}

type RegionModel struct {
	Code              types.String `tfsdk:"code"`
	Name              types.String `tfsdk:"name"`
	City              types.String `tfsdk:"city"`
	Country           types.String `tfsdk:"country"`
	Status            types.String `tfsdk:"status"`
	Available         types.Bool   `tfsdk:"available"`
	AvailableServices types.List   `tfsdk:"available_services"`
}

// regionsResponse is the body of GET /v1/regions.
type regionsResponse struct {
	Regions []regionResponse `json:"regions"`
}

type regionResponse struct {
	Code              string   `json:"code"`
	Name              string   `json:"name"`
	City              string   `json:"city"`
	Country           string   `json:"country"`
	Status            string   `json:"status"`
	AvailableServices []string `json:"available_services"`
}

func (d *RegionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

func (d *RegionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of WAYSCloud regions and the services available in each, from `GET /v1/regions`.",
		Attributes: map[string]schema.Attribute{
			"regions": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of regions, including regions that are not active.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Region code, lowercase (e.g., `no`, `se`, `dk`). This is the `region` that `wayscloud_app` and `wayscloud_redis_instance` take.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Region display name.",
						},
						"city": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "City where the region is located.",
						},
						"country": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "ISO 3166-1 alpha-2 code of the country where the region is located (e.g., `NO`).",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Region status (e.g., `active`, `maintenance`).",
						},
						"available": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the region is currently available, meaning its `status` is `active`.",
						},
						"available_services": schema.ListAttribute{
							Computed:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "Services that can currently be created in the region (e.g., `storage`, `databases`, `redis`, `apps`, `kubernetes`, `llm`). Empty unless the region is `active`.",
						},
					},
				},
			},
		},
	}
}

func (d *RegionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *RegionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading regions data source")

	respBody, err := d.client.Get(ctx, regionsPath)
	if err != nil {
		resp.Diagnostics.Append(dataSourceDiagnostic("wayscloud_regions", err)...)
		return
	}

	var body regionsResponse
	if err := json.Unmarshal(respBody, &body); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse regions response: %s", err))
		return
	}

	if len(body.Regions) == 0 {
		resp.Diagnostics.AddWarning("No Regions Returned", "The API returned no regions. This likely indicates a backend issue.")
	}

	var data RegionsDataSourceModel
	for _, r := range body.Regions {
		services := r.AvailableServices
		if services == nil {
			// An empty list rather than null, so contains() works on every region.
			services = []string{}
		}
		availableServices, diags := types.ListValueFrom(ctx, types.StringType, services)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.Regions = append(data.Regions, RegionModel{
			Code:              types.StringValue(r.Code),
			Name:              types.StringValue(r.Name),
			City:              types.StringValue(r.City),
			Country:           types.StringValue(r.Country),
			Status:            types.StringValue(r.Status),
			Available:         types.BoolValue(r.Status == regionStatusActive),
			AvailableServices: availableServices,
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
