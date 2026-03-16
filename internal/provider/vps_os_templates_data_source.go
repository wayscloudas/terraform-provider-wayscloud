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

var _ datasource.DataSource = &VPSOSTemplatesDataSource{}

func NewVPSOSTemplatesDataSource() datasource.DataSource {
	return &VPSOSTemplatesDataSource{}
}

type VPSOSTemplatesDataSource struct {
	client *client.Client
}

type VPSOSTemplatesDataSourceModel struct {
	Templates []VPSOSTemplateModel `tfsdk:"templates"`
}

type VPSOSTemplateModel struct {
	Slug     types.String `tfsdk:"slug"`
	Name     types.String `tfsdk:"name"`
	Family   types.String `tfsdk:"family"`
	Platform types.String `tfsdk:"platform"`
}

type vpsOSTemplateResponse struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Family   string `json:"family"`
	Platform string `json:"platform"`
}

func (d *VPSOSTemplatesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vps_os_templates"
}

func (d *VPSOSTemplatesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of available VPS operating system templates.",
		Attributes: map[string]schema.Attribute{
			"templates": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of available OS templates.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"slug": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Template slug for use in `wayscloud_vps` resource.",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Template display name.",
						},
						"family": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "OS family (e.g., `Ubuntu`, `Debian`, `Windows`).",
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

func (d *VPSOSTemplatesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *VPSOSTemplatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, "Reading VPS OS templates data source")

	respBody, err := d.client.Get(ctx, "/v1/vps/os-templates/")
	if err != nil {
		resp.Diagnostics.Append(dataSourceDiagnostic("wayscloud_vps_os_templates", err)...)
		return
	}

	var templates []vpsOSTemplateResponse
	if err := json.Unmarshal(respBody, &templates); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse OS templates response: %s", err))
		return
	}

	if len(templates) == 0 {
		resp.Diagnostics.AddWarning("No OS Templates Returned", "The API returned no OS templates. This likely indicates a backend issue.")
	}

	var data VPSOSTemplatesDataSourceModel
	for _, t := range templates {
		data.Templates = append(data.Templates, VPSOSTemplateModel{
			Slug:     types.StringValue(t.Slug),
			Name:     types.StringValue(t.Name),
			Family:   types.StringValue(t.Family),
			Platform: types.StringValue(t.Platform),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
