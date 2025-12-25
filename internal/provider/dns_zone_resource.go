// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DNSZoneResource{}
var _ resource.ResourceWithImportState = &DNSZoneResource{}

func NewDNSZoneResource() resource.Resource {
	return &DNSZoneResource{}
}

// DNSZoneResource defines the resource implementation.
type DNSZoneResource struct {
	client *client.Client
}

// DNSZoneResourceModel describes the resource data model.
type DNSZoneResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ZoneType     types.String `tfsdk:"zone_type"`
	Status       types.String `tfsdk:"status"`
	DNSSECEnabled types.Bool   `tfsdk:"dnssec_enabled"`
	RecordCount  types.Int64  `tfsdk:"record_count"`
	Nameservers  types.List   `tfsdk:"nameservers"`
	CreatedAt    types.String `tfsdk:"created_at"`
}

// API response/request structs
type dnsZoneCreateRequest struct {
	ZoneName string `json:"zone_name"`
	ZoneType string `json:"zone_type"`
}

type dnsZoneResponse struct {
	ZoneID        string   `json:"zone_id"`
	ZoneName      string   `json:"zone_name"`
	ZoneType      string   `json:"zone_type"`
	Status        string   `json:"status"`
	DNSSECEnabled bool     `json:"dnssec_enabled"`
	RecordCount   int64    `json:"record_count"`
	Nameservers   []string `json:"nameservers"`
	CreatedAt     string   `json:"created_at"`
}

func (r *DNSZoneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_zone"
}

func (r *DNSZoneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Manages a DNS zone in WAYSCloud.

A DNS zone represents a domain and contains DNS records. WAYSCloud provides authoritative DNS hosting with automatic replication to multiple nameservers.

## Example Usage

` + "```hcl" + `
resource "wayscloud_dns_zone" "example" {
  name = "example.com"
}
` + "```" + `

## Import

DNS zones can be imported using the zone name:

` + "```bash" + `
terraform import wayscloud_dns_zone.example example.com
` + "```" + `
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the DNS zone (UUID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The domain name for the DNS zone (e.g., `example.com`). Must be a valid FQDN.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Zone type: `master` (default), `slave`, or `parked`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Zone status: `active`, `suspended`, or `pending_deletion`.",
			},
			"dnssec_enabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether DNSSEC is enabled for this zone.",
			},
			"record_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of DNS records in this zone.",
			},
			"nameservers": schema.ListAttribute{
				Computed:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Authoritative nameservers for this zone. Configure these at your domain registrar.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the zone was created (ISO 8601).",
			},
		},
	}
}

func (r *DNSZoneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *DNSZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSZoneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default zone_type if not specified
	zoneType := "master"
	if !data.ZoneType.IsNull() && !data.ZoneType.IsUnknown() {
		zoneType = data.ZoneType.ValueString()
	}

	createReq := dnsZoneCreateRequest{
		ZoneName: data.Name.ValueString(),
		ZoneType: zoneType,
	}

	tflog.Debug(ctx, "Creating DNS zone", map[string]interface{}{
		"zone_name": createReq.ZoneName,
		"zone_type": createReq.ZoneType,
	})

	respBody, err := r.client.Post(ctx, "/v1/dns/zones", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create DNS zone: %s", err))
		return
	}

	var zone dnsZoneResponse
	if err := json.Unmarshal(respBody, &zone); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Map response to state
	data.ID = types.StringValue(zone.ZoneID)
	data.Name = types.StringValue(zone.ZoneName)
	data.ZoneType = types.StringValue(zone.ZoneType)
	data.Status = types.StringValue(zone.Status)
	data.DNSSECEnabled = types.BoolValue(zone.DNSSECEnabled)
	data.RecordCount = types.Int64Value(zone.RecordCount)
	data.CreatedAt = types.StringValue(zone.CreatedAt)

	// Convert nameservers to list
	nameservers, diags := types.ListValueFrom(ctx, types.StringType, zone.Nameservers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Nameservers = nameservers

	tflog.Trace(ctx, "Created DNS zone", map[string]interface{}{
		"zone_id":   zone.ZoneID,
		"zone_name": zone.ZoneName,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSZoneResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use zone name for API call (zones are identified by name in the API)
	zoneName := data.Name.ValueString()

	tflog.Debug(ctx, "Reading DNS zone", map[string]interface{}{
		"zone_name": zoneName,
	})

	respBody, err := r.client.Get(ctx, fmt.Sprintf("/v1/dns/zones/%s", zoneName))
	if err != nil {
		// Check if zone was deleted
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "DNS zone not found, removing from state", map[string]interface{}{
				"zone_name": zoneName,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read DNS zone: %s", err))
		return
	}

	var zone dnsZoneResponse
	if err := json.Unmarshal(respBody, &zone); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Map response to state
	data.ID = types.StringValue(zone.ZoneID)
	data.Name = types.StringValue(zone.ZoneName)
	data.ZoneType = types.StringValue(zone.ZoneType)
	data.Status = types.StringValue(zone.Status)
	data.DNSSECEnabled = types.BoolValue(zone.DNSSECEnabled)
	data.RecordCount = types.Int64Value(zone.RecordCount)
	data.CreatedAt = types.StringValue(zone.CreatedAt)

	// Convert nameservers to list
	nameservers, diags := types.ListValueFrom(ctx, types.StringType, zone.Nameservers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Nameservers = nameservers

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// DNS zones don't support updates (name and type require replacement)
	// Just read the current state
	var data DNSZoneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSZoneResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneName := data.Name.ValueString()

	tflog.Debug(ctx, "Deleting DNS zone", map[string]interface{}{
		"zone_name": zoneName,
	})

	_, err := r.client.Delete(ctx, fmt.Sprintf("/v1/dns/zones/%s", zoneName))
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "DNS zone already deleted", map[string]interface{}{
				"zone_name": zoneName,
			})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete DNS zone: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted DNS zone", map[string]interface{}{
		"zone_name": zoneName,
	})
}

func (r *DNSZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by zone name
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}
