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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &IoTDeviceResource{}
var _ resource.ResourceWithImportState = &IoTDeviceResource{}

func NewIoTDeviceResource() resource.Resource {
	return &IoTDeviceResource{}
}

// IoTDeviceResource defines the resource implementation.
type IoTDeviceResource struct {
	client *client.Client
}

// IoTDeviceResourceModel describes the resource data model.
type IoTDeviceResourceModel struct {
	DeviceID     types.String `tfsdk:"device_id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	DeviceType   types.String `tfsdk:"device_type"`
	Metadata     types.Map    `tfsdk:"metadata"`
	IsActive     types.Bool   `tfsdk:"is_active"`
	MQTTUsername types.String `tfsdk:"mqtt_username"`
	MQTTPassword types.String `tfsdk:"mqtt_password"`
}

// API request/response structs
type iotDeviceCreateRequest struct {
	DeviceID    string            `json:"device_id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	DeviceType  string            `json:"device_type,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	IsActive    bool              `json:"is_active"`
}

type iotDeviceUpdateRequest struct {
	Name        *string            `json:"name,omitempty"`
	Description *string            `json:"description,omitempty"`
	DeviceType  *string            `json:"device_type,omitempty"`
	Metadata    *map[string]string `json:"metadata,omitempty"`
	IsActive    *bool              `json:"is_active,omitempty"`
}

type iotDeviceResponse struct {
	DeviceID     string            `json:"device_id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	DeviceType   string            `json:"device_type"`
	Metadata     map[string]string `json:"metadata"`
	IsActive     bool              `json:"is_active"`
	MQTTUsername string            `json:"mqtt_username,omitempty"`
	MQTTPassword string            `json:"mqtt_password,omitempty"`
}

func (r *IoTDeviceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iot_device"
}

func (r *IoTDeviceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		MarkdownDescription: `
Manages an IoT device in WAYSCloud.

WAYSCloud IoT Platform provides device management with MQTT connectivity for telemetry and command delivery.

## Example Usage

` + "```hcl" + `
resource "wayscloud_iot_device" "sensor" {
  device_id   = "temp-sensor-01"
  name        = "Temperature Sensor #1"
  description = "Office temperature monitoring"
  device_type = "sensor"

  metadata = {
    location = "building-a"
    floor    = "2"
  }
}

output "mqtt_username" {
  value = wayscloud_iot_device.sensor.mqtt_username
}

output "mqtt_password" {
  value     = wayscloud_iot_device.sensor.mqtt_password
  sensitive = true
}
` + "```" + `

## Import

IoT devices can be imported using the device ID:

` + "```bash" + `
terraform import wayscloud_iot_device.sensor temp-sensor-01
` + "```" + `

~> **Note:** MQTT credentials cannot be retrieved after import.
`,

		Attributes: map[string]schema.Attribute{
			"device_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "User-defined unique device identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable device name.",
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Device description.",
			},
			"device_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Device type classification (e.g., `sensor`, `gateway`, `actuator`).",
			},
			"metadata": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Key-value metadata for the device.",
			},
			"is_active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the device is active. Default: `true`.",
			},
			"mqtt_username": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "MQTT username for device connectivity. Only available on initial creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mqtt_password": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "MQTT password for device connectivity. Only available on initial creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *IoTDeviceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IoTDeviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IoTDeviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := iotDeviceCreateRequest{
		DeviceID: data.DeviceID.ValueString(),
		Name:     data.Name.ValueString(),
		IsActive: data.IsActive.ValueBool(),
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		createReq.Description = data.Description.ValueString()
	}
	if !data.DeviceType.IsNull() && !data.DeviceType.IsUnknown() {
		createReq.DeviceType = data.DeviceType.ValueString()
	}
	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		metadata := make(map[string]string)
		resp.Diagnostics.Append(data.Metadata.ElementsAs(ctx, &metadata, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Metadata = metadata
	}

	tflog.Debug(ctx, "Creating IoT device", map[string]interface{}{
		"device_id": createReq.DeviceID,
		"name":      createReq.Name,
	})

	respBody, err := r.client.Post(ctx, "/v1/iot/devices", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create IoT device: %s", err))
		return
	}

	var device iotDeviceResponse
	if err := json.Unmarshal(respBody, &device); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	r.mapResponseToState(ctx, &data, &device)

	tflog.Trace(ctx, "Created IoT device", map[string]interface{}{
		"device_id": device.DeviceID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IoTDeviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IoTDeviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := data.DeviceID.ValueString()

	// Preserve MQTT credentials from state (not returned on read)
	mqttUsername := data.MQTTUsername
	mqttPassword := data.MQTTPassword

	tflog.Debug(ctx, "Reading IoT device", map[string]interface{}{
		"device_id": deviceID,
	})

	respBody, err := r.client.Get(ctx, fmt.Sprintf("/v1/iot/devices/%s", deviceID))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "IoT device not found, removing from state", map[string]interface{}{
				"device_id": deviceID,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read IoT device: %s", err))
		return
	}

	var device iotDeviceResponse
	if err := json.Unmarshal(respBody, &device); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	r.mapResponseToState(ctx, &data, &device)

	// Restore MQTT credentials from state
	data.MQTTUsername = mqttUsername
	data.MQTTPassword = mqttPassword

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IoTDeviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IoTDeviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state IoTDeviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := data.DeviceID.ValueString()

	updateReq := iotDeviceUpdateRequest{}
	hasChanges := false

	if data.Name.ValueString() != state.Name.ValueString() {
		name := data.Name.ValueString()
		updateReq.Name = &name
		hasChanges = true
	}
	if !data.Description.Equal(state.Description) {
		desc := data.Description.ValueString()
		updateReq.Description = &desc
		hasChanges = true
	}
	if !data.DeviceType.Equal(state.DeviceType) {
		dt := data.DeviceType.ValueString()
		updateReq.DeviceType = &dt
		hasChanges = true
	}
	if !data.Metadata.Equal(state.Metadata) {
		metadata := make(map[string]string)
		resp.Diagnostics.Append(data.Metadata.ElementsAs(ctx, &metadata, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.Metadata = &metadata
		hasChanges = true
	}
	if data.IsActive.ValueBool() != state.IsActive.ValueBool() {
		active := data.IsActive.ValueBool()
		updateReq.IsActive = &active
		hasChanges = true
	}

	if !hasChanges {
		tflog.Debug(ctx, "No changes detected for IoT device", map[string]interface{}{
			"device_id": deviceID,
		})
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	tflog.Debug(ctx, "Updating IoT device", map[string]interface{}{
		"device_id": deviceID,
	})

	respBody, err := r.client.Patch(ctx, fmt.Sprintf("/v1/iot/devices/%s", deviceID), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update IoT device: %s", err))
		return
	}

	var device iotDeviceResponse
	if err := json.Unmarshal(respBody, &device); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Preserve MQTT credentials from state
	mqttUsername := state.MQTTUsername
	mqttPassword := state.MQTTPassword

	r.mapResponseToState(ctx, &data, &device)

	data.MQTTUsername = mqttUsername
	data.MQTTPassword = mqttPassword

	tflog.Trace(ctx, "Updated IoT device", map[string]interface{}{
		"device_id": device.DeviceID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IoTDeviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IoTDeviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := data.DeviceID.ValueString()

	tflog.Debug(ctx, "Deleting IoT device", map[string]interface{}{
		"device_id": deviceID,
	})

	_, err := r.client.Delete(ctx, fmt.Sprintf("/v1/iot/devices/%s", deviceID))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "IoT device already deleted", map[string]interface{}{
				"device_id": deviceID,
			})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete IoT device: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted IoT device", map[string]interface{}{
		"device_id": deviceID,
	})
}

func (r *IoTDeviceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_id"), req.ID)...)
}

func (r *IoTDeviceResource) mapResponseToState(ctx context.Context, data *IoTDeviceResourceModel, device *iotDeviceResponse) {
	data.DeviceID = types.StringValue(device.DeviceID)
	data.Name = types.StringValue(device.Name)
	data.IsActive = types.BoolValue(device.IsActive)

	if device.Description != "" {
		data.Description = types.StringValue(device.Description)
	} else {
		data.Description = types.StringNull()
	}
	if device.DeviceType != "" {
		data.DeviceType = types.StringValue(device.DeviceType)
	} else {
		data.DeviceType = types.StringNull()
	}

	if len(device.Metadata) > 0 {
		metadataMap, diags := types.MapValueFrom(ctx, types.StringType, device.Metadata)
		if diags.HasError() {
			data.Metadata = types.MapNull(types.StringType)
		} else {
			data.Metadata = metadataMap
		}
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}

	if device.MQTTUsername != "" {
		data.MQTTUsername = types.StringValue(device.MQTTUsername)
	}
	if device.MQTTPassword != "" {
		data.MQTTPassword = types.StringValue(device.MQTTPassword)
	}
}
