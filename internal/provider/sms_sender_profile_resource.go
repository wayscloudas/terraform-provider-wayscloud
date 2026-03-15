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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SMSSenderProfileResource{}
var _ resource.ResourceWithImportState = &SMSSenderProfileResource{}

func NewSMSSenderProfileResource() resource.Resource {
	return &SMSSenderProfileResource{}
}

// SMSSenderProfileResource defines the resource implementation.
type SMSSenderProfileResource struct {
	client *client.Client
}

// SMSSenderProfileResourceModel describes the resource data model.
type SMSSenderProfileResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	SenderID       types.String `tfsdk:"sender_id"`
	AllowReply     types.Bool   `tfsdk:"allow_reply"`
	IsDefault      types.Bool   `tfsdk:"is_default"`
	ApprovalStatus types.String `tfsdk:"approval_status"`
}

// API request/response structs
type smsSenderProfileCreateRequest struct {
	Name       string `json:"name"`
	SenderID   string `json:"sender_id"`
	AllowReply bool   `json:"allow_reply"`
	IsDefault  bool   `json:"is_default"`
}

type smsSenderProfileResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	SenderID       string `json:"sender_id"`
	AllowReply     bool   `json:"allow_reply"`
	IsDefault      bool   `json:"is_default"`
	ApprovalStatus string `json:"approval_status"`
}

type smsSenderProfileListResponse struct {
	Profiles []smsSenderProfileResponse `json:"profiles"`
}

func (r *SMSSenderProfileResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sms_sender_profile"
}

func (r *SMSSenderProfileResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		MarkdownDescription: `
Manages an SMS sender profile in WAYSCloud.

Sender profiles define the sender ID (name or number) displayed when sending SMS messages.

## Example Usage

` + "```hcl" + `
resource "wayscloud_sms_sender_profile" "alerts" {
  name       = "Alert System"
  sender_id  = "WAYSCloud"
  allow_reply = false
  is_default  = true
}
` + "```" + `

## Import

SMS sender profiles can be imported using the profile UUID:

` + "```bash" + `
terraform import wayscloud_sms_sender_profile.alerts 550e8400-e29b-41d4-a716-446655440000
` + "```" + `

~> **Note:** Sender profiles do not support updates. Any change forces recreation.
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the sender profile (UUID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Profile name for internal reference.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sender_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Sender ID displayed to recipients (alphanumeric name or phone number).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"allow_reply": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether recipients can reply to messages. Default: `true`.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"is_default": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether this is the default sender profile. Default: `false`.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"approval_status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Approval status: `pending`, `approved`, `rejected`.",
			},
		},
	}
}

func (r *SMSSenderProfileResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SMSSenderProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SMSSenderProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := smsSenderProfileCreateRequest{
		Name:       data.Name.ValueString(),
		SenderID:   data.SenderID.ValueString(),
		AllowReply: data.AllowReply.ValueBool(),
		IsDefault:  data.IsDefault.ValueBool(),
	}

	tflog.Debug(ctx, "Creating SMS sender profile", map[string]interface{}{
		"name":      createReq.Name,
		"sender_id": createReq.SenderID,
	})

	respBody, err := r.client.Post(ctx, "/v1/sms/sender-profiles", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create SMS sender profile: %s", err))
		return
	}

	var profile smsSenderProfileResponse
	if err := json.Unmarshal(respBody, &profile); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	r.mapResponseToState(&data, &profile)

	tflog.Trace(ctx, "Created SMS sender profile", map[string]interface{}{
		"id": profile.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SMSSenderProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SMSSenderProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	profileID := data.ID.ValueString()

	tflog.Debug(ctx, "Reading SMS sender profile", map[string]interface{}{
		"id": profileID,
	})

	// API only supports list endpoint; fetch all and filter
	respBody, err := r.client.Get(ctx, "/v1/sms/sender-profiles")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read SMS sender profiles: %s", err))
		return
	}

	var listResp smsSenderProfileListResponse
	if err := json.Unmarshal(respBody, &listResp); err != nil {
		// Try unmarshalling as array directly
		var profiles []smsSenderProfileResponse
		if err2 := json.Unmarshal(respBody, &profiles); err2 != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
			return
		}
		listResp.Profiles = profiles
	}

	// Find our profile by ID
	var found *smsSenderProfileResponse
	for _, p := range listResp.Profiles {
		if p.ID == profileID {
			found = &p
			break
		}
	}

	if found == nil {
		tflog.Debug(ctx, "SMS sender profile not found, removing from state", map[string]interface{}{
			"id": profileID,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	r.mapResponseToState(&data, found)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SMSSenderProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// SMS sender profiles don't support updates (all fields require replacement)
	var data SMSSenderProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SMSSenderProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SMSSenderProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	profileID := data.ID.ValueString()

	tflog.Debug(ctx, "Deleting SMS sender profile", map[string]interface{}{
		"id": profileID,
	})

	_, err := r.client.Delete(ctx, fmt.Sprintf("/v1/sms/sender-profiles/%s", profileID))
	if err != nil {
		if is404(err) {
			tflog.Debug(ctx, "SMS sender profile already deleted", map[string]interface{}{
				"id": profileID,
			})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete SMS sender profile: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted SMS sender profile", map[string]interface{}{
		"id": profileID,
	})
}

func (r *SMSSenderProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SMSSenderProfileResource) mapResponseToState(data *SMSSenderProfileResourceModel, profile *smsSenderProfileResponse) {
	data.ID = types.StringValue(profile.ID)
	data.Name = types.StringValue(profile.Name)
	data.SenderID = types.StringValue(profile.SenderID)
	data.AllowReply = types.BoolValue(profile.AllowReply)
	data.IsDefault = types.BoolValue(profile.IsDefault)
	data.ApprovalStatus = types.StringValue(profile.ApprovalStatus)
}
