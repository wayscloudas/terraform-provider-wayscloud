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

var _ datasource.DataSource = &DatabaseTypesDataSource{}

func NewDatabaseTypesDataSource() datasource.DataSource {
	return &DatabaseTypesDataSource{}
}

type DatabaseTypesDataSource struct {
	client *client.Client
}

type DatabaseTypesDataSourceModel struct {
	DatabaseTypes []DatabaseTypeModel `tfsdk:"database_types"`
}

type DatabaseTypeModel struct {
	Tier         types.String `tfsdk:"tier"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	IsEncrypted  types.Bool   `tfsdk:"is_encrypted"`
}

type databaseTypeResponse struct {
	Tier        string `json:"tier"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsEncrypted bool   `json:"is_encrypted"`
}

func (d *DatabaseTypesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_types"
}

func (d *DatabaseTypesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of available WAYSCloud database types and tiers.",
		Attributes: map[string]schema.Attribute{
			"database_types": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of available database types.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"tier": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Tier level (e.g., `standard`, `encrypted`).",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Display name of the database tier.",
						},
						"description": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Human-readable description of the database tier.",
						},
						"is_encrypted": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether this tier uses encryption at rest.",
						},
					},
				},
			},
		},
	}
}

func (d *DatabaseTypesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DatabaseTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading database types data source")

	respBody, err := d.client.Get(ctx, "/api/v1/dashboard/databases/tiers")
	if err != nil {
		resp.Diagnostics.Append(dataSourceDiagnostic("wayscloud_database_types", err)...)
		return
	}

	var dbTypes []databaseTypeResponse
	if err := json.Unmarshal(respBody, &dbTypes); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse database types response: %s", err))
		return
	}

	if len(dbTypes) == 0 {
		resp.Diagnostics.AddWarning("No Database Types Returned", "The API returned no database types. This likely indicates a backend issue.")
	}

	var data DatabaseTypesDataSourceModel
	for _, dt := range dbTypes {
		data.DatabaseTypes = append(data.DatabaseTypes, DatabaseTypeModel{
			Tier:        types.StringValue(dt.Tier),
			Name:        types.StringValue(dt.Name),
			Description: types.StringValue(dt.Description),
			IsEncrypted: types.BoolValue(dt.IsEncrypted),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
