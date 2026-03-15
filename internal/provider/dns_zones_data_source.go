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

var _ datasource.DataSource = &DNSZonesDataSource{}

func NewDNSZonesDataSource() datasource.DataSource {
	return &DNSZonesDataSource{}
}

type DNSZonesDataSource struct {
	client *client.Client
}

type DNSZonesDataSourceModel struct {
	Zones []DNSZoneDataModel `tfsdk:"zones"`
}

type DNSZoneDataModel struct {
	Name   types.String `tfsdk:"name"`
	Status types.String `tfsdk:"status"`
}

type dnsZoneListResponse struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (d *DNSZonesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_zones"
}

func (d *DNSZonesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of DNS zones in the account.",
		Attributes: map[string]schema.Attribute{
			"zones": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of DNS zones.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Zone name (domain).",
						},
						"status": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Zone status.",
						},
					},
				},
			},
		},
	}
}

func (d *DNSZonesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DNSZonesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading DNS zones data source")

	respBody, err := d.client.Get(ctx, "/v1/dns/zones")
	if err != nil {
		resp.Diagnostics.Append(dataSourceDiagnostic("wayscloud_dns_zones", err)...)
		return
	}

	var zones []dnsZoneListResponse
	if err := json.Unmarshal(respBody, &zones); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse DNS zones response: %s", err))
		return
	}

	// No empty warning for dns_zones — an account may legitimately have no zones

	var data DNSZonesDataSourceModel
	for _, z := range zones {
		data.Zones = append(data.Zones, DNSZoneDataModel{
			Name:   types.StringValue(z.Name),
			Status: types.StringValue(z.Status),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
