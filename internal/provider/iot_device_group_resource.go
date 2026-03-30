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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &IoTDeviceGroupResource{}
var _ resource.ResourceWithImportState = &IoTDeviceGroupResource{}

func NewIoTDeviceGroupResource() resource.Resource {
	return &IoTDeviceGroupResource{}
}

// IoTDeviceGroupResource defines the resource implementation.
type IoTDeviceGroupResource struct {
	client *client.Client
}

// IoTDeviceGroupResourceModel describes the resource data model.
type IoTDeviceGroupResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	DeviceCount types.Int64  `tfsdk:"device_count"`
	CreatedAt   types.String `tfsdk:"created_at"`
}

// API request/response structs
type iotGroupCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type iotGroupUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type iotGroupResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DeviceCount int64  `json:"device_count"`
	CreatedAt   string `json:"created_at"`
}

func (r *IoTDeviceGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iot_device_group"
}

func (r *IoTDeviceGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		MarkdownDescription: `
Manages an IoT device group in WAYSCloud.

Device groups allow you to organize IoT devices into logical collections for bulk operations, monitoring rules, and access control.

## Example Usage

` + "```hcl" + `
resource "wayscloud_iot_device_group" "sensors" {
  name        = "Temperature Sensors"
  description = "All temperature sensors in building A"
}
` + "```" + `

## Import

IoT device groups can be imported using the group ID:

` + "```bash" + `
terraform import wayscloud_iot_device_group.sensors 550e8400-e29b-41d4-a716-446655440000
` + "```" + `
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the device group (UUID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable name for the device group.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description of the device group.",
			},
			"device_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Number of devices in the group.",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the group was created (ISO 8601).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *IoTDeviceGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IoTDeviceGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IoTDeviceGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := iotGroupCreateRequest{
		Name: data.Name.ValueString(),
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		createReq.Description = data.Description.ValueString()
	}

	tflog.Debug(ctx, "Creating IoT device group", map[string]interface{}{
		"name": createReq.Name,
	})

	respBody, err := r.client.Post(ctx, "/v1/iot/groups", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create IoT device group: %s", err))
		return
	}

	var group iotGroupResponse
	if err := json.Unmarshal(respBody, &group); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	r.mapResponseToState(&data, &group)

	tflog.Trace(ctx, "Created IoT device group", map[string]interface{}{
		"group_id": group.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IoTDeviceGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IoTDeviceGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := data.ID.ValueString()

	tflog.Debug(ctx, "Reading IoT device group", map[string]interface{}{
		"group_id": groupID,
	})

	respBody, err := r.client.Get(ctx, fmt.Sprintf("/v1/iot/groups/%s", groupID))
	if err != nil {
		if is404(err) {
			tflog.Debug(ctx, "IoT device group not found, removing from state", map[string]interface{}{
				"group_id": groupID,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read IoT device group: %s", err))
		return
	}

	var group iotGroupResponse
	if err := json.Unmarshal(respBody, &group); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	r.mapResponseToState(&data, &group)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IoTDeviceGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IoTDeviceGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state IoTDeviceGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := state.ID.ValueString()

	updateReq := iotGroupUpdateRequest{}
	hasChanges := false

	if data.Name.ValueString() != state.Name.ValueString() {
		updateReq.Name = data.Name.ValueString()
		hasChanges = true
	}
	if !data.Description.Equal(state.Description) {
		updateReq.Description = data.Description.ValueString()
		hasChanges = true
	}

	if !hasChanges {
		tflog.Debug(ctx, "No changes detected for IoT device group", map[string]interface{}{
			"group_id": groupID,
		})
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	tflog.Debug(ctx, "Updating IoT device group", map[string]interface{}{
		"group_id": groupID,
	})

	respBody, err := r.client.Put(ctx, fmt.Sprintf("/v1/iot/groups/%s", groupID), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update IoT device group: %s", err))
		return
	}

	var group iotGroupResponse
	if err := json.Unmarshal(respBody, &group); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	r.mapResponseToState(&data, &group)

	tflog.Trace(ctx, "Updated IoT device group", map[string]interface{}{
		"group_id": group.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IoTDeviceGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IoTDeviceGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groupID := data.ID.ValueString()

	tflog.Debug(ctx, "Deleting IoT device group", map[string]interface{}{
		"group_id": groupID,
	})

	_, err := r.client.Delete(ctx, fmt.Sprintf("/v1/iot/groups/%s", groupID))
	if err != nil {
		// Ignore 404 errors (already deleted)
		if is404(err) {
			tflog.Debug(ctx, "IoT device group already deleted", map[string]interface{}{
				"group_id": groupID,
			})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete IoT device group: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted IoT device group", map[string]interface{}{
		"group_id": groupID,
	})
}

func (r *IoTDeviceGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapResponseToState maps the API response to the Terraform state model
func (r *IoTDeviceGroupResource) mapResponseToState(data *IoTDeviceGroupResourceModel, group *iotGroupResponse) {
	data.ID = types.StringValue(group.ID)
	data.Name = types.StringValue(group.Name)
	data.DeviceCount = types.Int64Value(group.DeviceCount)
	data.CreatedAt = types.StringValue(group.CreatedAt)

	if group.Description != "" {
		data.Description = types.StringValue(group.Description)
	} else {
		data.Description = types.StringNull()
	}
}
