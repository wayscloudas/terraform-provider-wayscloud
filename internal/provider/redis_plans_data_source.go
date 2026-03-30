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

var _ datasource.DataSource = &RedisPlansDataSource{}

func NewRedisPlansDataSource() datasource.DataSource {
	return &RedisPlansDataSource{}
}

type RedisPlansDataSource struct {
	client *client.Client
}

type RedisPlansDataSourceModel struct {
	Plans []RedisPlanModel `tfsdk:"plans"`
}

type RedisPlanModel struct {
	ID           types.String  `tfsdk:"id"`
	Name         types.String  `tfsdk:"name"`
	MemoryMB     types.Int64   `tfsdk:"memory_mb"`
	MonthlyPrice types.Float64 `tfsdk:"monthly_price"`
	Currency     types.String  `tfsdk:"currency"`
}

type redisPlanResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	MemoryMB     int64   `json:"memory_mb"`
	MonthlyPrice float64 `json:"monthly_price"`
	Currency     string  `json:"currency"`
}

func (d *RedisPlansDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redis_plans"
}

func (d *RedisPlansDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of available WAYSCloud Redis plans.",
		Attributes: map[string]schema.Attribute{
			"plans": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of available Redis plans.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Plan identifier.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Plan display name.",
						},
						"memory_mb": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "Memory allocation in megabytes.",
						},
						"monthly_price": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "Monthly price for the plan.",
						},
						"currency": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Price currency (e.g., `NOK`, `EUR`).",
						},
					},
				},
			},
		},
	}
}

func (d *RedisPlansDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RedisPlansDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading Redis plans data source")

	respBody, err := d.client.Get(ctx, "/v1/redis/plans")
	if err != nil {
		resp.Diagnostics.Append(dataSourceDiagnostic("wayscloud_redis_plans", err)...)
		return
	}

	var plans []redisPlanResponse
	if err := json.Unmarshal(respBody, &plans); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse Redis plans response: %s", err))
		return
	}

	if len(plans) == 0 {
		resp.Diagnostics.AddWarning("No Redis Plans Returned", "The API returned no Redis plans. This likely indicates a backend issue.")
	}

	var data RedisPlansDataSourceModel
	for _, p := range plans {
		data.Plans = append(data.Plans, RedisPlanModel{
			ID:           types.StringValue(p.ID),
			Name:         types.StringValue(p.Name),
			MemoryMB:     types.Int64Value(p.MemoryMB),
			MonthlyPrice: types.Float64Value(p.MonthlyPrice),
			Currency:     types.StringValue(p.Currency),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
