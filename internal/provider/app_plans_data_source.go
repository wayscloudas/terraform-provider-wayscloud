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

var _ datasource.DataSource = &AppPlansDataSource{}

func NewAppPlansDataSource() datasource.DataSource {
	return &AppPlansDataSource{}
}

type AppPlansDataSource struct {
	client *client.Client
}

type AppPlansDataSourceModel struct {
	Plans []AppPlanModel `tfsdk:"plans"`
}

type AppPlanModel struct {
	ID              types.String  `tfsdk:"id"`
	Name            types.String  `tfsdk:"name"`
	VCPU            types.Float64 `tfsdk:"vcpu"`
	RAMMB           types.Int64   `tfsdk:"ram_mb"`
	MonthlyPriceNOK types.Float64 `tfsdk:"monthly_price_nok"`
}

type appPlanResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	VCPU            float64 `json:"vcpu"`
	RAMMB           int64   `json:"ram_mb"`
	MonthlyPriceNOK float64 `json:"monthly_price_nok"`
}

func (d *AppPlansDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app_plans"
}

func (d *AppPlansDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of available App Platform plans.",
		Attributes: map[string]schema.Attribute{
			"plans": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of available app plans.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Plan ID for use in `wayscloud_app` resource.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Plan display name.",
						},
						"vcpu": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "Number of virtual CPUs.",
						},
						"ram_mb": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "RAM in megabytes.",
						},
						"monthly_price_nok": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "Monthly price in NOK.",
						},
					},
				},
			},
		},
	}
}

func (d *AppPlansDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AppPlansDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading app plans data source")

	respBody, err := d.client.Get(ctx, "/v1/apps/plans")
	if err != nil {
		resp.Diagnostics.Append(dataSourceDiagnostic("wayscloud_app_plans", err)...)
		return
	}

	var plans []appPlanResponse
	if err := json.Unmarshal(respBody, &plans); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse app plans response: %s", err))
		return
	}

	if len(plans) == 0 {
		resp.Diagnostics.AddWarning("No App Plans Returned", "The API returned no app plans. This likely indicates a backend issue.")
	}

	var data AppPlansDataSourceModel
	for _, p := range plans {
		data.Plans = append(data.Plans, AppPlanModel{
			ID:              types.StringValue(p.ID),
			Name:            types.StringValue(p.Name),
			VCPU:            types.Float64Value(p.VCPU),
			RAMMB:           types.Int64Value(p.RAMMB),
			MonthlyPriceNOK: types.Float64Value(p.MonthlyPriceNOK),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
