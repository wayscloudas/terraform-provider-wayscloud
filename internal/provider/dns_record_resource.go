// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DNSRecordResource{}
var _ resource.ResourceWithImportState = &DNSRecordResource{}

func NewDNSRecordResource() resource.Resource {
	return &DNSRecordResource{}
}

// DNSRecordResource defines the resource implementation.
type DNSRecordResource struct {
	client *client.Client
}

// DNSRecordResourceModel describes the resource data model.
type DNSRecordResourceModel struct {
	ID         types.String `tfsdk:"id"`
	ZoneName   types.String `tfsdk:"zone_name"`
	ZoneID     types.String `tfsdk:"zone_id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Value      types.String `tfsdk:"value"`
	TTL        types.Int64  `tfsdk:"ttl"`
	Priority   types.Int64  `tfsdk:"priority"`
	Weight     types.Int64  `tfsdk:"weight"`
	Port       types.Int64  `tfsdk:"port"`
	Status     types.String `tfsdk:"status"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

// API request/response structs
type dnsRecordCreateRequest struct {
	RecordType string `json:"record_type"`
	Host       string `json:"host"`
	Record     string `json:"record"`
	TTL        int64  `json:"ttl"`
	Priority   *int64 `json:"priority,omitempty"`
	Weight     *int64 `json:"weight,omitempty"`
	Port       *int64 `json:"port,omitempty"`
}

type dnsRecordUpdateRequest struct {
	Host     *string `json:"host,omitempty"`
	Record   *string `json:"record,omitempty"`
	TTL      *int64  `json:"ttl,omitempty"`
	Priority *int64  `json:"priority,omitempty"`
	Weight   *int64  `json:"weight,omitempty"`
	Port     *int64  `json:"port,omitempty"`
}

type dnsRecordResponse struct {
	RecordID   string `json:"record_id"`
	ZoneID     string `json:"zone_id"`
	RecordType string `json:"record_type"`
	Host       string `json:"host"`
	Record     string `json:"record"`
	TTL        int64  `json:"ttl"`
	Priority   *int64 `json:"priority,omitempty"`
	Weight     *int64 `json:"weight,omitempty"`
	Port       *int64 `json:"port,omitempty"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}

func (r *DNSRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *DNSRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Manages a DNS record in WAYSCloud.

DNS records define how your domain responds to DNS queries. Supports all common record types including A, AAAA, CNAME, MX, TXT, and SRV.

## Example Usage

### A Record

` + "```hcl" + `
resource "wayscloud_dns_record" "www" {
  zone_name = wayscloud_dns_zone.example.name
  name      = "www"
  type      = "A"
  value     = "192.0.2.1"
  ttl       = 300
}
` + "```" + `

### MX Record

` + "```hcl" + `
resource "wayscloud_dns_record" "mail" {
  zone_name = wayscloud_dns_zone.example.name
  name      = ""
  type      = "MX"
  value     = "mail.example.com"
  ttl       = 3600
  priority  = 10
}
` + "```" + `

### CNAME Record

` + "```hcl" + `
resource "wayscloud_dns_record" "blog" {
  zone_name = wayscloud_dns_zone.example.name
  name      = "blog"
  type      = "CNAME"
  value     = "example.com"
  ttl       = 3600
}
` + "```" + `

### TXT Record (SPF)

` + "```hcl" + `
resource "wayscloud_dns_record" "spf" {
  zone_name = wayscloud_dns_zone.example.name
  name      = ""
  type      = "TXT"
  value     = "v=spf1 include:_spf.wayscloud.services ~all"
  ttl       = 3600
}
` + "```" + `

## Import

DNS records can be imported using the format ` + "`zone_name/record_id`" + `:

` + "```bash" + `
terraform import wayscloud_dns_record.www example.com/550e8400-e29b-41d4-a716-446655440000
` + "```" + `
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the DNS record (UUID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The DNS zone name this record belongs to (e.g., `example.com`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"zone_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The zone UUID this record belongs to.",
			},
			"name": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				MarkdownDescription: "Hostname or subdomain. Use empty string for root domain (`@`), or `*` for wildcard.",
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "DNS record type: `A`, `AAAA`, `CNAME`, `MX`, `TXT`, `NS`, `SRV`, `CAA`, `PTR`, `SPF`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Record value (IP address, hostname, or text depending on record type).",
			},
			"ttl": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(3600),
				MarkdownDescription: "Time to live in seconds (60-2592000). Default: 3600.",
			},
			"priority": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Priority for MX and SRV records. Lower values have higher priority.",
			},
			"weight": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Weight for SRV records. Used for load balancing between records with same priority.",
			},
			"port": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Port number for SRV records.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Record status: `active`, `suspended`.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the record was created (ISO 8601).",
			},
		},
	}
}

func (r *DNSRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := dnsRecordCreateRequest{
		RecordType: data.Type.ValueString(),
		Host:       data.Name.ValueString(),
		Record:     data.Value.ValueString(),
		TTL:        data.TTL.ValueInt64(),
	}

	// Add optional fields for MX/SRV records
	if !data.Priority.IsNull() && !data.Priority.IsUnknown() {
		priority := data.Priority.ValueInt64()
		createReq.Priority = &priority
	}
	if !data.Weight.IsNull() && !data.Weight.IsUnknown() {
		weight := data.Weight.ValueInt64()
		createReq.Weight = &weight
	}
	if !data.Port.IsNull() && !data.Port.IsUnknown() {
		port := data.Port.ValueInt64()
		createReq.Port = &port
	}

	zoneName := data.ZoneName.ValueString()

	tflog.Debug(ctx, "Creating DNS record", map[string]interface{}{
		"zone_name":   zoneName,
		"record_type": createReq.RecordType,
		"host":        createReq.Host,
	})

	respBody, err := r.client.Post(ctx, fmt.Sprintf("/v1/dns/zones/%s/records", zoneName), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create DNS record: %s", err))
		return
	}

	var record dnsRecordResponse
	if err := json.Unmarshal(respBody, &record); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Map response to state
	r.mapResponseToState(&data, &record, zoneName)

	tflog.Trace(ctx, "Created DNS record", map[string]interface{}{
		"record_id": record.RecordID,
		"zone_name": zoneName,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneName := data.ZoneName.ValueString()
	recordID := data.ID.ValueString()

	tflog.Debug(ctx, "Reading DNS record", map[string]interface{}{
		"zone_name": zoneName,
		"record_id": recordID,
	})

	respBody, err := r.client.Get(ctx, fmt.Sprintf("/v1/dns/zones/%s/records/%s", zoneName, recordID))
	if err != nil {
		// Check if record was deleted
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "DNS record not found, removing from state", map[string]interface{}{
				"record_id": recordID,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read DNS record: %s", err))
		return
	}

	var record dnsRecordResponse
	if err := json.Unmarshal(respBody, &record); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Map response to state
	r.mapResponseToState(&data, &record, zoneName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSRecordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DNSRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneName := data.ZoneName.ValueString()
	recordID := state.ID.ValueString()

	updateReq := dnsRecordUpdateRequest{}
	hasChanges := false

	// Only include changed fields
	if data.Name.ValueString() != state.Name.ValueString() {
		name := data.Name.ValueString()
		updateReq.Host = &name
		hasChanges = true
	}
	if data.Value.ValueString() != state.Value.ValueString() {
		value := data.Value.ValueString()
		updateReq.Record = &value
		hasChanges = true
	}
	if data.TTL.ValueInt64() != state.TTL.ValueInt64() {
		ttl := data.TTL.ValueInt64()
		updateReq.TTL = &ttl
		hasChanges = true
	}
	if !data.Priority.Equal(state.Priority) {
		if !data.Priority.IsNull() && !data.Priority.IsUnknown() {
			priority := data.Priority.ValueInt64()
			updateReq.Priority = &priority
		}
		hasChanges = true
	}
	if !data.Weight.Equal(state.Weight) {
		if !data.Weight.IsNull() && !data.Weight.IsUnknown() {
			weight := data.Weight.ValueInt64()
			updateReq.Weight = &weight
		}
		hasChanges = true
	}
	if !data.Port.Equal(state.Port) {
		if !data.Port.IsNull() && !data.Port.IsUnknown() {
			port := data.Port.ValueInt64()
			updateReq.Port = &port
		}
		hasChanges = true
	}

	if !hasChanges {
		tflog.Debug(ctx, "No changes detected for DNS record", map[string]interface{}{
			"record_id": recordID,
		})
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	tflog.Debug(ctx, "Updating DNS record", map[string]interface{}{
		"zone_name": zoneName,
		"record_id": recordID,
	})

	respBody, err := r.client.Patch(ctx, fmt.Sprintf("/v1/dns/zones/%s/records/%s", zoneName, recordID), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update DNS record: %s", err))
		return
	}

	var record dnsRecordResponse
	if err := json.Unmarshal(respBody, &record); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Map response to state
	r.mapResponseToState(&data, &record, zoneName)

	tflog.Trace(ctx, "Updated DNS record", map[string]interface{}{
		"record_id": record.RecordID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSRecordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneName := data.ZoneName.ValueString()
	recordID := data.ID.ValueString()

	tflog.Debug(ctx, "Deleting DNS record", map[string]interface{}{
		"zone_name": zoneName,
		"record_id": recordID,
	})

	_, err := r.client.Delete(ctx, fmt.Sprintf("/v1/dns/zones/%s/records/%s", zoneName, recordID))
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "DNS record already deleted", map[string]interface{}{
				"record_id": recordID,
			})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete DNS record: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted DNS record", map[string]interface{}{
		"record_id": recordID,
	})
}

func (r *DNSRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: zone_name/record_id
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: zone_name/record_id, got: %s", req.ID),
		)
		return
	}

	zoneName := parts[0]
	recordID := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone_name"), zoneName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), recordID)...)
}

// mapResponseToState maps the API response to the Terraform state model
func (r *DNSRecordResource) mapResponseToState(data *DNSRecordResourceModel, record *dnsRecordResponse, zoneName string) {
	data.ID = types.StringValue(record.RecordID)
	data.ZoneID = types.StringValue(record.ZoneID)
	data.ZoneName = types.StringValue(zoneName)
	data.Name = types.StringValue(record.Host)
	data.Type = types.StringValue(record.RecordType)
	data.Value = types.StringValue(record.Record)
	data.TTL = types.Int64Value(record.TTL)
	data.Status = types.StringValue(record.Status)
	data.CreatedAt = types.StringValue(record.CreatedAt)

	// Handle optional fields
	if record.Priority != nil {
		data.Priority = types.Int64Value(*record.Priority)
	} else {
		data.Priority = types.Int64Null()
	}
	if record.Weight != nil {
		data.Weight = types.Int64Value(*record.Weight)
	} else {
		data.Weight = types.Int64Null()
	}
	if record.Port != nil {
		data.Port = types.Int64Value(*record.Port)
	} else {
		data.Port = types.Int64Null()
	}
}
