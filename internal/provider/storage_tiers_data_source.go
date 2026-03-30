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

var _ datasource.DataSource = &StorageTiersDataSource{}

func NewStorageTiersDataSource() datasource.DataSource {
	return &StorageTiersDataSource{}
}

type StorageTiersDataSource struct {
	client *client.Client
}

type StorageTiersDataSourceModel struct {
	Tiers []StorageTierModel `tfsdk:"tiers"`
}

type StorageTierModel struct {
	ID           types.String  `tfsdk:"id"`
	Name         types.String  `tfsdk:"name"`
	Description  types.String  `tfsdk:"description"`
	MonthlyPrice types.Float64 `tfsdk:"monthly_price"`
	Currency     types.String  `tfsdk:"currency"`
}

type storageTierResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	MonthlyPrice float64 `json:"monthly_price"`
	Currency     string  `json:"currency"`
}

func (d *StorageTiersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_tiers"
}

func (d *StorageTiersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of available WAYSCloud S3 storage tiers.",
		Attributes: map[string]schema.Attribute{
			"tiers": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of available storage tiers.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Tier identifier.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Tier display name.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Human-readable description of the storage tier.",
						},
						"monthly_price": schema.Float64Attribute{
							Computed:            true,
							MarkdownDescription: "Monthly price for the tier.",
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

func (d *StorageTiersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StorageTiersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading storage tiers data source")

	respBody, err := d.client.Get(ctx, "/api/v1/dashboard/storage/tiers")
	if err != nil {
		resp.Diagnostics.Append(dataSourceDiagnostic("wayscloud_storage_tiers", err)...)
		return
	}

	var tiers []storageTierResponse
	if err := json.Unmarshal(respBody, &tiers); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse storage tiers response: %s", err))
		return
	}

	if len(tiers) == 0 {
		resp.Diagnostics.AddWarning("No Storage Tiers Returned", "The API returned no storage tiers. This likely indicates a backend issue.")
	}

	var data StorageTiersDataSourceModel
	for _, t := range tiers {
		data.Tiers = append(data.Tiers, StorageTierModel{
			ID:           types.StringValue(t.ID),
			Name:         types.StringValue(t.Name),
			Description:  types.StringValue(t.Description),
			MonthlyPrice: types.Float64Value(t.MonthlyPrice),
			Currency:     types.StringValue(t.Currency),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
