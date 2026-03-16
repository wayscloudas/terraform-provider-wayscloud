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
	PlanCode     types.String  `tfsdk:"plan_code"`
	Name         types.String  `tfsdk:"name"`
	CPUCores     types.Int64   `tfsdk:"cpu_cores"`
	RAMMB        types.Int64   `tfsdk:"ram_mb"`
	DiskGB       types.Int64   `tfsdk:"disk_gb"`
	MonthlyPrice types.Float64 `tfsdk:"monthly_price"`
	Currency     types.String  `tfsdk:"currency"`
	Region       types.String  `tfsdk:"region"`
	Platform     types.String  `tfsdk:"platform"`
}

type vpsPlanResponse struct {
	SKU          string  `json:"sku"`
	PlanCode     *string `json:"plan_code"`
	Name         string  `json:"name"`
	Provider     string  `json:"provider"`
	CPUCores     int64   `json:"cpu_cores"`
	RAMMB        int64   `json:"ram_mb"`
	DiskGB       int64   `json:"disk_gb"`
	MonthlyPrice float64 `json:"monthly_price"`
	Currency     string  `json:"currency"`
	IsActive     bool    `json:"is_active"`
}

func (d *VPSPlansDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vps_plans"
}

func (d *VPSPlansDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of available VPS plans. Prices are returned in the customer's preferred currency.",
		Attributes: map[string]schema.Attribute{
			"plans": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of available VPS plans.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"plan_code": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Plan code for use in `wayscloud_vps` resource.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Plan display name.",
						},
						"cpu_cores": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of CPU cores.",
						},
						"ram_mb": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "RAM in megabytes.",
						},
						"disk_gb": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Disk size in gigabytes.",
						},
						"monthly_price": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "Monthly price in the customer's preferred currency.",
						},
						"currency": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Currency code (e.g., `NOK`, `SEK`, `DKK`, `EUR`).",
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

	respBody, err := d.client.Get(ctx, "/v1/vps/plans/")
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
		planCode := ""
		if p.PlanCode != nil {
			planCode = *p.PlanCode
		} else {
			planCode = p.SKU
		}
		data.Plans = append(data.Plans, VPSPlanModel{
			PlanCode:     types.StringValue(planCode),
			Name:         types.StringValue(p.Name),
			CPUCores:     types.Int64Value(p.CPUCores),
			RAMMB:        types.Int64Value(p.RAMMB),
			DiskGB:       types.Int64Value(p.DiskGB),
			MonthlyPrice: types.Float64Value(p.MonthlyPrice),
			Currency:     types.StringValue(p.Currency),
			Region:       types.StringValue(""),
			Platform:     types.StringValue(""),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
