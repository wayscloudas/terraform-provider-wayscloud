// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VPSResource{}
var _ resource.ResourceWithImportState = &VPSResource{}

func NewVPSResource() resource.Resource {
	return &VPSResource{}
}

// VPSResource defines the resource implementation.
type VPSResource struct {
	client *client.Client
}

// VPSResourceModel describes the resource data model.
type VPSResourceModel struct {
	ID              types.String  `tfsdk:"id"`
	ProviderVMID    types.String  `tfsdk:"provider_vm_id"`
	Hostname        types.String  `tfsdk:"hostname"`
	DisplayName     types.String  `tfsdk:"display_name"`
	PlanCode        types.String  `tfsdk:"plan_code"`
	Region          types.String  `tfsdk:"region"`
	OSTemplate      types.String  `tfsdk:"os_template"`
	SSHKeys         types.List    `tfsdk:"ssh_keys"`
	Status          types.String  `tfsdk:"status"`
	PowerState      types.String  `tfsdk:"power_state"`
	IPv4Address     types.String  `tfsdk:"ipv4_address"`
	IPv6Address     types.String  `tfsdk:"ipv6_address"`
	VCPU            types.Int64   `tfsdk:"vcpu"`
	RAMMB           types.Int64   `tfsdk:"ram_mb"`
	DiskGB          types.Int64   `tfsdk:"disk_gb"`
	MonthlyPriceNOK types.Float64 `tfsdk:"monthly_price_nok"`
	CreatedAt       types.String  `tfsdk:"created_at"`
	ProvisionedAt   types.String  `tfsdk:"provisioned_at"`
}

// API request/response structs
type vpsCreateRequest struct {
	Hostname    string   `json:"hostname"`
	PlanCode    string   `json:"plan_code"`
	Region      string   `json:"region"`
	OSTemplate  string   `json:"os_template"`
	DisplayName string   `json:"display_name,omitempty"`
	SSHKeys     []string `json:"ssh_keys,omitempty"`
}

type vpsResponse struct {
	ID              string   `json:"id"`
	ProviderVMID    string   `json:"provider_vm_id"`
	Hostname        string   `json:"hostname"`
	DisplayName     *string  `json:"display_name,omitempty"`
	PlanCode        string   `json:"plan_code"`
	PlanName        *string  `json:"plan_name,omitempty"`
	Region          string   `json:"region"`
	OSTemplate      *string  `json:"os_template,omitempty"`
	Status          string   `json:"status"`
	PowerState      string   `json:"power_state"`
	IPv4Address     *string  `json:"ipv4_address,omitempty"`
	IPv6Address     *string  `json:"ipv6_address,omitempty"`
	VCPU            *int64   `json:"vcpu,omitempty"`
	RAMMB           *int64   `json:"ram_mb,omitempty"`
	DiskGB          *int64   `json:"disk_gb,omitempty"`
	MonthlyPriceNOK *float64 `json:"monthly_price_nok,omitempty"`
	CreatedAt       string   `json:"created_at"`
	ProvisionedAt   *string  `json:"provisioned_at,omitempty"`
}

func (r *VPSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vps"
}

func (r *VPSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Manages a VPS (Virtual Private Server) instance in WAYSCloud.

WAYSCloud VPS provides fully managed virtual servers with automatic provisioning,
monitoring, and optional SSH key injection.

## Example Usage

### Basic VPS with Ubuntu

` + "```hcl" + `
resource "wayscloud_vps" "web" {
  hostname    = "web01.example.com"
  display_name = "Production Web Server"
  plan_code   = "NO-Start-Linux-2cpu-4096mb-30gb"
  region      = "NO"
  os_template = "ubuntu-22.04"

  ssh_keys = [
    "ssh-rsa AAAAB3NzaC1yc2E..."
  ]
}

output "server_ip" {
  value = wayscloud_vps.web.ipv4_address
}
` + "```" + `

### Windows Server

` + "```hcl" + `
resource "wayscloud_vps" "win" {
  hostname    = "win01.example.com"
  plan_code   = "NO-Premium-Windows-4cpu-8192mb-100gb"
  region      = "NO"
  os_template = "windows-server-2022"
}
` + "```" + `

## Import

VPS instances can be imported using the VPS ID:

` + "```bash" + `
terraform import wayscloud_vps.web 550e8400-e29b-41d4-a716-446655440000
` + "```" + `

## Provisioning Notes

After creation, the VPS enters ` + "`provisioning`" + ` status. Terraform will poll
until the status changes to ` + "`active`" + ` (typically 2-5 minutes).

SSH keys are injected via cloud-init during initial boot.
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the VPS instance (UUID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"provider_vm_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Provider-specific VM ID (internal use).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "VPS hostname (FQDN). Example: `web01.example.com`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "User-friendly display name for the VPS.",
			},
			"plan_code": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "VPS plan code. Use the `wayscloud_vps_plans` data source to list available plans.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Datacenter region. Example: `NO` (Norway).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"os_template": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "OS template. Example: `ubuntu-22.04`, `debian-12`, `windows-server-2022`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ssh_keys": schema.ListAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "List of SSH public keys for access (Linux only).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "VPS status: `provisioning`, `active`, `stopped`, `terminated`, `error`.",
			},
			"power_state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Power state: `on`, `off`.",
			},
			"ipv4_address": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Primary IPv4 address.",
			},
			"ipv6_address": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Primary IPv6 address (if available).",
			},
			"vcpu": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of virtual CPUs.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"ram_mb": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "RAM in megabytes.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"disk_gb": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Disk size in gigabytes.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"monthly_price_nok": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "Monthly price in NOK.",
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the VPS was created (ISO 8601).",
			},
			"provisioned_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the VPS finished provisioning (ISO 8601).",
			},
		},
	}
}

func (r *VPSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VPSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VPSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := vpsCreateRequest{
		Hostname:   data.Hostname.ValueString(),
		PlanCode:   data.PlanCode.ValueString(),
		Region:     data.Region.ValueString(),
		OSTemplate: data.OSTemplate.ValueString(),
	}

	if !data.DisplayName.IsNull() {
		createReq.DisplayName = data.DisplayName.ValueString()
	}

	// Extract SSH keys from list
	if !data.SSHKeys.IsNull() {
		var sshKeys []string
		resp.Diagnostics.Append(data.SSHKeys.ElementsAs(ctx, &sshKeys, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.SSHKeys = sshKeys
	}

	tflog.Debug(ctx, "Creating VPS", map[string]interface{}{
		"hostname":   createReq.Hostname,
		"plan_code":  createReq.PlanCode,
		"region":     createReq.Region,
		"os_template": createReq.OSTemplate,
	})

	respBody, err := r.client.Post(ctx, "/v1/vps/", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create VPS: %s", err))
		return
	}

	var vps vpsResponse
	if err := json.Unmarshal(respBody, &vps); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Map response to state
	r.mapResponseToState(&data, &vps)

	tflog.Trace(ctx, "Created VPS", map[string]interface{}{
		"id":       vps.ID,
		"hostname": vps.Hostname,
		"status":   vps.Status,
	})

	// Save initial state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Wait for VPS to be ready
	if vps.Status == "provisioning" {
		tflog.Debug(ctx, "Waiting for VPS to be ready", map[string]interface{}{
			"id": vps.ID,
		})

		readyVPS, err := r.waitForReady(ctx, vps.ID)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"VPS Still Provisioning",
				fmt.Sprintf("VPS created but not yet ready. Status: %s. Run terraform refresh to update.", vps.Status),
			)
			return
		}

		// Update with final state
		r.mapResponseToState(&data, readyVPS)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func (r *VPSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VPSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vpsID := data.ID.ValueString()

	tflog.Debug(ctx, "Reading VPS", map[string]interface{}{
		"id": vpsID,
	})

	respBody, err := r.client.Get(ctx, fmt.Sprintf("/v1/vps/%s", vpsID))
	if err != nil {
		// Check if VPS was deleted
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "VPS not found, removing from state", map[string]interface{}{
				"id": vpsID,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read VPS: %s", err))
		return
	}

	var vps vpsResponse
	if err := json.Unmarshal(respBody, &vps); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Map response to state, preserving SSH keys from state (not returned in response)
	sshKeys := data.SSHKeys
	r.mapResponseToState(&data, &vps)
	data.SSHKeys = sshKeys

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// VPS update is limited - most fields require replacement
	// Only display_name can be updated in-place (if API supports it)
	var data VPSResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// For now, just save the state - actual update would require API support
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VPSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VPSResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vpsID := data.ID.ValueString()

	tflog.Debug(ctx, "Deleting VPS", map[string]interface{}{
		"id": vpsID,
	})

	_, err := r.client.Delete(ctx, fmt.Sprintf("/v1/vps/%s", vpsID))
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "VPS already deleted", map[string]interface{}{
				"id": vpsID,
			})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete VPS: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted VPS", map[string]interface{}{
		"id": vpsID,
	})
}

func (r *VPSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapResponseToState maps the API response to the Terraform state model
func (r *VPSResource) mapResponseToState(data *VPSResourceModel, vps *vpsResponse) {
	data.ID = types.StringValue(vps.ID)
	data.ProviderVMID = types.StringValue(vps.ProviderVMID)
	data.Hostname = types.StringValue(vps.Hostname)
	data.PlanCode = types.StringValue(vps.PlanCode)
	data.Region = types.StringValue(vps.Region)
	data.Status = types.StringValue(vps.Status)
	data.PowerState = types.StringValue(vps.PowerState)
	data.CreatedAt = types.StringValue(vps.CreatedAt)

	// Handle optional fields
	if vps.DisplayName != nil {
		data.DisplayName = types.StringValue(*vps.DisplayName)
	}
	if vps.OSTemplate != nil {
		data.OSTemplate = types.StringValue(*vps.OSTemplate)
	}
	if vps.IPv4Address != nil {
		data.IPv4Address = types.StringValue(*vps.IPv4Address)
	} else {
		data.IPv4Address = types.StringNull()
	}
	if vps.IPv6Address != nil {
		data.IPv6Address = types.StringValue(*vps.IPv6Address)
	} else {
		data.IPv6Address = types.StringNull()
	}
	if vps.VCPU != nil {
		data.VCPU = types.Int64Value(*vps.VCPU)
	}
	if vps.RAMMB != nil {
		data.RAMMB = types.Int64Value(*vps.RAMMB)
	}
	if vps.DiskGB != nil {
		data.DiskGB = types.Int64Value(*vps.DiskGB)
	}
	if vps.MonthlyPriceNOK != nil {
		data.MonthlyPriceNOK = types.Float64Value(*vps.MonthlyPriceNOK)
	}
	if vps.ProvisionedAt != nil {
		data.ProvisionedAt = types.StringValue(*vps.ProvisionedAt)
	}
}

// waitForReady polls the VPS until it's ready or timeout
func (r *VPSResource) waitForReady(ctx context.Context, vpsID string) (*vpsResponse, error) {
	timeout := 10 * time.Minute
	pollInterval := 15 * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		respBody, err := r.client.Get(ctx, fmt.Sprintf("/v1/vps/%s", vpsID))
		if err != nil {
			return nil, err
		}

		var vps vpsResponse
		if err := json.Unmarshal(respBody, &vps); err != nil {
			return nil, err
		}

		tflog.Debug(ctx, "Polling VPS status", map[string]interface{}{
			"id":     vpsID,
			"status": vps.Status,
		})

		switch vps.Status {
		case "active":
			return &vps, nil
		case "error", "terminated":
			return nil, fmt.Errorf("VPS entered %s state", vps.Status)
		case "provisioning":
			// Continue polling
		default:
			// Unknown status, continue polling
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
			// Continue
		}
	}

	return nil, fmt.Errorf("timeout waiting for VPS to be ready")
}
