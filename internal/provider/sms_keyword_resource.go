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
var _ resource.Resource = &SMSKeywordResource{}
var _ resource.ResourceWithImportState = &SMSKeywordResource{}

func NewSMSKeywordResource() resource.Resource {
	return &SMSKeywordResource{}
}

// SMSKeywordResource defines the resource implementation.
type SMSKeywordResource struct {
	client *client.Client
}

// SMSKeywordResourceModel describes the resource data model.
type SMSKeywordResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Keyword          types.String `tfsdk:"keyword"`
	Description      types.String `tfsdk:"description"`
	WebhookURL       types.String `tfsdk:"webhook_url"`
	AutoReplyEnabled types.Bool   `tfsdk:"auto_reply_enabled"`
	AutoReplyMessage types.String `tfsdk:"auto_reply_message"`
	IsActive         types.Bool   `tfsdk:"is_active"`
}

// API request/response structs
type smsKeywordCreateRequest struct {
	Keyword          string `json:"keyword"`
	Description      string `json:"description,omitempty"`
	WebhookURL       string `json:"webhook_url,omitempty"`
	AutoReplyEnabled bool   `json:"auto_reply_enabled"`
	AutoReplyMessage string `json:"auto_reply_message,omitempty"`
	IsActive         bool   `json:"is_active"`
}

type smsKeywordUpdateRequest struct {
	Description      *string `json:"description,omitempty"`
	WebhookURL       *string `json:"webhook_url,omitempty"`
	AutoReplyEnabled *bool   `json:"auto_reply_enabled,omitempty"`
	AutoReplyMessage *string `json:"auto_reply_message,omitempty"`
	IsActive         *bool   `json:"is_active,omitempty"`
}

type smsKeywordResponse struct {
	ID               string `json:"id"`
	Keyword          string `json:"keyword"`
	Description      string `json:"description"`
	WebhookURL       string `json:"webhook_url"`
	AutoReplyEnabled bool   `json:"auto_reply_enabled"`
	AutoReplyMessage string `json:"auto_reply_message"`
	IsActive         bool   `json:"is_active"`
}

func (r *SMSKeywordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sms_keyword"
}

func (r *SMSKeywordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		MarkdownDescription: `
Manages an SMS keyword in WAYSCloud.

Keywords allow you to receive and respond to inbound SMS messages matching specific words.

## Example Usage

` + "```hcl" + `
resource "wayscloud_sms_keyword" "help" {
  keyword            = "HELP"
  description        = "Help keyword for customer support"
  webhook_url        = "https://api.example.com/sms/inbound"
  auto_reply_enabled = true
  auto_reply_message = "Thank you for contacting us. We will respond shortly."
}
` + "```" + `

## Import

SMS keywords can be imported using the keyword UUID:

` + "```bash" + `
terraform import wayscloud_sms_keyword.help 550e8400-e29b-41d4-a716-446655440000
` + "```" + `
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the keyword (UUID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"keyword": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The keyword to match in inbound SMS messages.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Description of the keyword's purpose.",
			},
			"webhook_url": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "URL to receive webhook notifications when this keyword is matched.",
			},
			"auto_reply_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Whether to automatically reply when keyword is matched. Default: `false`.",
			},
			"auto_reply_message": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Message to send as automatic reply (requires `auto_reply_enabled = true`).",
			},
			"is_active": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Whether the keyword is active. Default: `true`.",
			},
		},
	}
}

func (r *SMSKeywordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SMSKeywordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SMSKeywordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := smsKeywordCreateRequest{
		Keyword:          data.Keyword.ValueString(),
		AutoReplyEnabled: data.AutoReplyEnabled.ValueBool(),
		IsActive:         data.IsActive.ValueBool(),
	}

	if !data.Description.IsNull() && !data.Description.IsUnknown() {
		createReq.Description = data.Description.ValueString()
	}
	if !data.WebhookURL.IsNull() && !data.WebhookURL.IsUnknown() {
		createReq.WebhookURL = data.WebhookURL.ValueString()
	}
	if !data.AutoReplyMessage.IsNull() && !data.AutoReplyMessage.IsUnknown() {
		createReq.AutoReplyMessage = data.AutoReplyMessage.ValueString()
	}

	tflog.Debug(ctx, "Creating SMS keyword", map[string]interface{}{
		"keyword": createReq.Keyword,
	})

	respBody, err := r.client.Post(ctx, "/v1/sms/keywords", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create SMS keyword: %s", err))
		return
	}

	var keyword smsKeywordResponse
	if err := json.Unmarshal(respBody, &keyword); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	r.mapResponseToState(&data, &keyword)

	tflog.Trace(ctx, "Created SMS keyword", map[string]interface{}{
		"id":      keyword.ID,
		"keyword": keyword.Keyword,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SMSKeywordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SMSKeywordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keywordID := data.ID.ValueString()

	tflog.Debug(ctx, "Reading SMS keyword", map[string]interface{}{
		"id": keywordID,
	})

	respBody, err := r.client.Get(ctx, fmt.Sprintf("/v1/sms/keywords/%s", keywordID))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "SMS keyword not found, removing from state", map[string]interface{}{
				"id": keywordID,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read SMS keyword: %s", err))
		return
	}

	var keyword smsKeywordResponse
	if err := json.Unmarshal(respBody, &keyword); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	r.mapResponseToState(&data, &keyword)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SMSKeywordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SMSKeywordResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state SMSKeywordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keywordID := state.ID.ValueString()

	updateReq := smsKeywordUpdateRequest{}
	hasChanges := false

	if !data.Description.Equal(state.Description) {
		desc := data.Description.ValueString()
		updateReq.Description = &desc
		hasChanges = true
	}
	if !data.WebhookURL.Equal(state.WebhookURL) {
		url := data.WebhookURL.ValueString()
		updateReq.WebhookURL = &url
		hasChanges = true
	}
	if data.AutoReplyEnabled.ValueBool() != state.AutoReplyEnabled.ValueBool() {
		enabled := data.AutoReplyEnabled.ValueBool()
		updateReq.AutoReplyEnabled = &enabled
		hasChanges = true
	}
	if !data.AutoReplyMessage.Equal(state.AutoReplyMessage) {
		msg := data.AutoReplyMessage.ValueString()
		updateReq.AutoReplyMessage = &msg
		hasChanges = true
	}
	if data.IsActive.ValueBool() != state.IsActive.ValueBool() {
		active := data.IsActive.ValueBool()
		updateReq.IsActive = &active
		hasChanges = true
	}

	if !hasChanges {
		tflog.Debug(ctx, "No changes detected for SMS keyword", map[string]interface{}{
			"id": keywordID,
		})
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	tflog.Debug(ctx, "Updating SMS keyword", map[string]interface{}{
		"id": keywordID,
	})

	respBody, err := r.client.Patch(ctx, fmt.Sprintf("/v1/sms/keywords/%s", keywordID), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update SMS keyword: %s", err))
		return
	}

	var keyword smsKeywordResponse
	if err := json.Unmarshal(respBody, &keyword); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	r.mapResponseToState(&data, &keyword)

	tflog.Trace(ctx, "Updated SMS keyword", map[string]interface{}{
		"id":      keyword.ID,
		"keyword": keyword.Keyword,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SMSKeywordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SMSKeywordResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	keywordID := data.ID.ValueString()

	tflog.Debug(ctx, "Deleting SMS keyword", map[string]interface{}{
		"id": keywordID,
	})

	_, err := r.client.Delete(ctx, fmt.Sprintf("/v1/sms/keywords/%s", keywordID))
	if err != nil {
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "SMS keyword already deleted", map[string]interface{}{
				"id": keywordID,
			})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete SMS keyword: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted SMS keyword", map[string]interface{}{
		"id": keywordID,
	})
}

func (r *SMSKeywordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *SMSKeywordResource) mapResponseToState(data *SMSKeywordResourceModel, keyword *smsKeywordResponse) {
	data.ID = types.StringValue(keyword.ID)
	data.Keyword = types.StringValue(keyword.Keyword)
	data.AutoReplyEnabled = types.BoolValue(keyword.AutoReplyEnabled)
	data.IsActive = types.BoolValue(keyword.IsActive)

	if keyword.Description != "" {
		data.Description = types.StringValue(keyword.Description)
	} else {
		data.Description = types.StringNull()
	}
	if keyword.WebhookURL != "" {
		data.WebhookURL = types.StringValue(keyword.WebhookURL)
	} else {
		data.WebhookURL = types.StringNull()
	}
	if keyword.AutoReplyMessage != "" {
		data.AutoReplyMessage = types.StringValue(keyword.AutoReplyMessage)
	} else {
		data.AutoReplyMessage = types.StringNull()
	}
}
