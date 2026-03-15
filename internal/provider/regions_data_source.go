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
	Code        types.String `tfsdk:"code"`
	Name        types.String `tfsdk:"name"`
	Country     types.String `tfsdk:"country"`
	Available   types.Bool   `tfsdk:"available"`
}

type regionResponse struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	Country   string `json:"country"`
	Available bool   `json:"available"`
}

func (d *RegionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

func (d *RegionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of available WAYSCloud regions.",
		Attributes: map[string]schema.Attribute{
			"regions": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of available regions.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Region code (e.g., `no`, `eu`).",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Region display name.",
						},
						"country": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Country where the region is located.",
						},
						"available": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the region is currently available.",
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

	respBody, err := d.client.Get(ctx, "/v1/regions")
	if err != nil {
		resp.Diagnostics.Append(dataSourceDiagnostic("wayscloud_regions", err)...)
		return
	}

	var regions []regionResponse
	if err := json.Unmarshal(respBody, &regions); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse regions response: %s", err))
		return
	}

	if len(regions) == 0 {
		resp.Diagnostics.AddWarning("No Regions Returned", "The API returned no regions. This likely indicates a backend issue.")
	}

	var data RegionsDataSourceModel
	for _, r := range regions {
		data.Regions = append(data.Regions, RegionModel{
			Code:      types.StringValue(r.Code),
			Name:      types.StringValue(r.Name),
			Country:   types.StringValue(r.Country),
			Available: types.BoolValue(r.Available),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
