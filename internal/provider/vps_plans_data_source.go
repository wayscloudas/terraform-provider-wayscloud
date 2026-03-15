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

var _ datasource.DataSource = &VPSPlansDataSource{}

func NewVPSPlansDataSource() datasource.DataSource {
	return &VPSPlansDataSource{}
}

type VPSPlansDataSource struct {
	client *client.Client
}

type VPSPlansDataSourceModel struct {
	Plans []VPSPlanModel `tfsdk:"plans"`
}

type VPSPlanModel struct {
	Code            types.String  `tfsdk:"code"`
	Name            types.String  `tfsdk:"name"`
	VCPU            types.Int64   `tfsdk:"vcpu"`
	RAMMB           types.Int64   `tfsdk:"ram_mb"`
	DiskGB          types.Int64   `tfsdk:"disk_gb"`
	MonthlyPriceNOK types.Float64 `tfsdk:"monthly_price_nok"`
	Region          types.String  `tfsdk:"region"`
	Platform        types.String  `tfsdk:"platform"`
}

type vpsPlanResponse struct {
	Code            string  `json:"code"`
	Name            string  `json:"name"`
	VCPU            int64   `json:"vcpu"`
	RAMMB           int64   `json:"ram_mb"`
	DiskGB          int64   `json:"disk_gb"`
	MonthlyPriceNOK float64 `json:"monthly_price_nok"`
	Region          string  `json:"region"`
	Platform        string  `json:"platform"`
}

func (d *VPSPlansDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vps_plans"
}

func (d *VPSPlansDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of available VPS plans.",
		Attributes: map[string]schema.Attribute{
			"plans": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of available VPS plans.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"code": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Plan code for use in `wayscloud_vps` resource.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Plan display name.",
						},
						"vcpu": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of virtual CPUs.",
						},
						"ram_mb": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "RAM in megabytes.",
						},
						"disk_gb": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Disk size in gigabytes.",
						},
						"monthly_price_nok": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "Monthly price in NOK.",
						},
						"region": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Region code.",
						},
						"platform": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Platform type: `linux` or `windows`.",
						},
					},
				},
			},
		},
	}
}

func (d *VPSPlansDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VPSPlansDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading VPS plans data source")

	respBody, err := d.client.Get(ctx, "/v1/vps/plans")
	if err != nil {
		resp.Diagnostics.Append(dataSourceDiagnostic("wayscloud_vps_plans", err)...)
		return
	}

	var plans []vpsPlanResponse
	if err := json.Unmarshal(respBody, &plans); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse VPS plans response: %s", err))
		return
	}

	if len(plans) == 0 {
		resp.Diagnostics.AddWarning("No VPS Plans Returned", "The API returned no VPS plans. This likely indicates a backend issue.")
	}

	var data VPSPlansDataSourceModel
	for _, p := range plans {
		data.Plans = append(data.Plans, VPSPlanModel{
			Code:            types.StringValue(p.Code),
			Name:            types.StringValue(p.Name),
			VCPU:            types.Int64Value(p.VCPU),
			RAMMB:           types.Int64Value(p.RAMMB),
			DiskGB:          types.Int64Value(p.DiskGB),
			MonthlyPriceNOK: types.Float64Value(p.MonthlyPriceNOK),
			Region:          types.StringValue(p.Region),
			Platform:        types.StringValue(p.Platform),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
