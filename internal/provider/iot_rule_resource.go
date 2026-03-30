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
var _ resource.Resource = &IoTRuleResource{}
var _ resource.ResourceWithImportState = &IoTRuleResource{}

func NewIoTRuleResource() resource.Resource {
	return &IoTRuleResource{}
}

// IoTRuleResource defines the resource implementation.
type IoTRuleResource struct {
	client *client.Client
}

// IoTRuleResourceModel describes the resource data model.
type IoTRuleResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	RuleType        types.String `tfsdk:"rule_type"`
	ScopeType       types.String `tfsdk:"scope_type"`
	Severity        types.String `tfsdk:"severity"`
	IsEnabled       types.Bool   `tfsdk:"is_enabled"`
	CooldownSeconds types.Int64  `tfsdk:"cooldown_seconds"`
	AutoResolve     types.Bool   `tfsdk:"auto_resolve"`
	Config          types.String `tfsdk:"config"`
	CreatedAt       types.String `tfsdk:"created_at"`
}

// API request/response structs
type iotRuleCreateRequest struct {
	Name            string          `json:"name"`
	RuleType        string          `json:"rule_type"`
	ScopeType       string          `json:"scope_type"`
	Severity        string          `json:"severity"`
	IsEnabled       bool            `json:"is_enabled"`
	CooldownSeconds int64           `json:"cooldown_seconds,omitempty"`
	AutoResolve     bool            `json:"auto_resolve"`
	Config          json.RawMessage `json:"config"`
}

type iotRuleUpdateRequest struct {
	Name            string          `json:"name,omitempty"`
	RuleType        string          `json:"rule_type,omitempty"`
	ScopeType       string          `json:"scope_type,omitempty"`
	Severity        string          `json:"severity,omitempty"`
	IsEnabled       *bool           `json:"is_enabled,omitempty"`
	CooldownSeconds *int64          `json:"cooldown_seconds,omitempty"`
	AutoResolve     *bool           `json:"auto_resolve,omitempty"`
	Config          json.RawMessage `json:"config,omitempty"`
}

type iotRuleResponse struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	RuleType        string          `json:"rule_type"`
	ScopeType       string          `json:"scope_type"`
	Severity        string          `json:"severity"`
	IsEnabled       bool            `json:"is_enabled"`
	CooldownSeconds int64           `json:"cooldown_seconds"`
	AutoResolve     bool            `json:"auto_resolve"`
	Config          json.RawMessage `json:"config"`
	CreatedAt       string          `json:"created_at"`
}

func (r *IoTRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_iot_rule"
}

func (r *IoTRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		MarkdownDescription: `
Manages an IoT alerting rule in WAYSCloud.

IoT rules define conditions that trigger alerts based on device telemetry data. Rules can target an entire fleet, a device group, a device profile, or a single device.

## Example Usage

` + "```hcl" + `
resource "wayscloud_iot_rule" "temp_alert" {
  name      = "High Temperature Alert"
  rule_type = "threshold"
  scope_type = "fleet"
  severity   = "warning"

  config = jsonencode({
    field     = "temperature"
    operator  = ">"
    value     = 80
    duration  = 300
  })

  cooldown_seconds = 600
  auto_resolve     = true
}
` + "```" + `

### Missing Data Rule

` + "```hcl" + `
resource "wayscloud_iot_rule" "offline_check" {
  name       = "Device Offline Detection"
  rule_type  = "missing_data"
  scope_type = "group"
  severity   = "critical"

  config = jsonencode({
    timeout_seconds = 3600
    group_id        = wayscloud_iot_device_group.sensors.id
  })

  auto_resolve = true
}
` + "```" + `

## Import

IoT rules can be imported using the rule ID:

` + "```bash" + `
terraform import wayscloud_iot_rule.temp_alert 550e8400-e29b-41d4-a716-446655440000
` + "```" + `
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the rule (UUID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable name for the rule.",
			},
			"rule_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Type of rule: `missing_data`, `offline`, `threshold`, `message_rate`, `reconnect_rate`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scope_type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Scope of the rule: `fleet`, `group`, `profile`, `device`.",
			},
			"severity": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Alert severity: `critical`, `warning`, `info`.",
			},
			"is_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the rule is enabled. Default: `true`.",
			},
			"cooldown_seconds": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Minimum seconds between repeated alerts for the same rule. Prevents alert fatigue.",
			},
			"auto_resolve": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether alerts auto-resolve when the condition clears. Default: `false`.",
			},
			"config": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Rule configuration as a JSON string. Structure depends on `rule_type`. Use `jsonencode()` to build.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the rule was created (ISO 8601).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *IoTRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IoTRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IoTRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := iotRuleCreateRequest{
		Name:        data.Name.ValueString(),
		RuleType:    data.RuleType.ValueString(),
		ScopeType:   data.ScopeType.ValueString(),
		Severity:    data.Severity.ValueString(),
		IsEnabled:   data.IsEnabled.ValueBool(),
		AutoResolve: data.AutoResolve.ValueBool(),
		Config:      json.RawMessage(data.Config.ValueString()),
	}

	if !data.CooldownSeconds.IsNull() && !data.CooldownSeconds.IsUnknown() {
		createReq.CooldownSeconds = data.CooldownSeconds.ValueInt64()
	}

	tflog.Debug(ctx, "Creating IoT rule", map[string]interface{}{
		"name":      createReq.Name,
		"rule_type": createReq.RuleType,
	})

	respBody, err := r.client.Post(ctx, "/v1/iot/rules", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create IoT rule: %s", err))
		return
	}

	var rule iotRuleResponse
	if err := json.Unmarshal(respBody, &rule); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	r.mapResponseToState(&data, &rule)

	tflog.Trace(ctx, "Created IoT rule", map[string]interface{}{
		"rule_id": rule.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IoTRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IoTRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleID := data.ID.ValueString()

	tflog.Debug(ctx, "Reading IoT rule", map[string]interface{}{
		"rule_id": ruleID,
	})

	respBody, err := r.client.Get(ctx, fmt.Sprintf("/v1/iot/rules/%s", ruleID))
	if err != nil {
		if is404(err) {
			tflog.Debug(ctx, "IoT rule not found, removing from state", map[string]interface{}{
				"rule_id": ruleID,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read IoT rule: %s", err))
		return
	}

	var rule iotRuleResponse
	if err := json.Unmarshal(respBody, &rule); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	r.mapResponseToState(&data, &rule)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IoTRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IoTRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state IoTRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleID := state.ID.ValueString()

	updateReq := iotRuleUpdateRequest{}
	hasChanges := false

	if data.Name.ValueString() != state.Name.ValueString() {
		updateReq.Name = data.Name.ValueString()
		hasChanges = true
	}
	if data.ScopeType.ValueString() != state.ScopeType.ValueString() {
		updateReq.ScopeType = data.ScopeType.ValueString()
		hasChanges = true
	}
	if data.Severity.ValueString() != state.Severity.ValueString() {
		updateReq.Severity = data.Severity.ValueString()
		hasChanges = true
	}
	if data.IsEnabled.ValueBool() != state.IsEnabled.ValueBool() {
		enabled := data.IsEnabled.ValueBool()
		updateReq.IsEnabled = &enabled
		hasChanges = true
	}
	if !data.CooldownSeconds.Equal(state.CooldownSeconds) {
		if !data.CooldownSeconds.IsNull() && !data.CooldownSeconds.IsUnknown() {
			cooldown := data.CooldownSeconds.ValueInt64()
			updateReq.CooldownSeconds = &cooldown
		}
		hasChanges = true
	}
	if data.AutoResolve.ValueBool() != state.AutoResolve.ValueBool() {
		autoResolve := data.AutoResolve.ValueBool()
		updateReq.AutoResolve = &autoResolve
		hasChanges = true
	}
	if data.Config.ValueString() != state.Config.ValueString() {
		updateReq.Config = json.RawMessage(data.Config.ValueString())
		hasChanges = true
	}

	if !hasChanges {
		tflog.Debug(ctx, "No changes detected for IoT rule", map[string]interface{}{
			"rule_id": ruleID,
		})
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	tflog.Debug(ctx, "Updating IoT rule", map[string]interface{}{
		"rule_id": ruleID,
	})

	respBody, err := r.client.Put(ctx, fmt.Sprintf("/v1/iot/rules/%s", ruleID), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update IoT rule: %s", err))
		return
	}

	var rule iotRuleResponse
	if err := json.Unmarshal(respBody, &rule); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	r.mapResponseToState(&data, &rule)

	tflog.Trace(ctx, "Updated IoT rule", map[string]interface{}{
		"rule_id": rule.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IoTRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IoTRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ruleID := data.ID.ValueString()

	tflog.Debug(ctx, "Deleting IoT rule", map[string]interface{}{
		"rule_id": ruleID,
	})

	_, err := r.client.Delete(ctx, fmt.Sprintf("/v1/iot/rules/%s", ruleID))
	if err != nil {
		// Ignore 404 errors (already deleted)
		if is404(err) {
			tflog.Debug(ctx, "IoT rule already deleted", map[string]interface{}{
				"rule_id": ruleID,
			})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete IoT rule: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted IoT rule", map[string]interface{}{
		"rule_id": ruleID,
	})
}

func (r *IoTRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapResponseToState maps the API response to the Terraform state model
func (r *IoTRuleResource) mapResponseToState(data *IoTRuleResourceModel, rule *iotRuleResponse) {
	data.ID = types.StringValue(rule.ID)
	data.Name = types.StringValue(rule.Name)
	data.RuleType = types.StringValue(rule.RuleType)
	data.ScopeType = types.StringValue(rule.ScopeType)
	data.Severity = types.StringValue(rule.Severity)
	data.IsEnabled = types.BoolValue(rule.IsEnabled)
	data.AutoResolve = types.BoolValue(rule.AutoResolve)
	data.CreatedAt = types.StringValue(rule.CreatedAt)

	if rule.CooldownSeconds > 0 {
		data.CooldownSeconds = types.Int64Value(rule.CooldownSeconds)
	} else {
		data.CooldownSeconds = types.Int64Null()
	}

	if rule.Config != nil {
		data.Config = types.StringValue(string(rule.Config))
	}
}
