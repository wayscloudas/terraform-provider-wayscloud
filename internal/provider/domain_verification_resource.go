// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"strings"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DomainVerificationResource{}
var _ resource.ResourceWithImportState = &DomainVerificationResource{}

func NewDomainVerificationResource() resource.Resource {
	return &DomainVerificationResource{}
}

// DomainVerificationResource defines the resource implementation.
type DomainVerificationResource struct {
	client *client.Client
}

// DomainVerificationResourceModel describes the resource data model.
type DomainVerificationResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Domain             types.String `tfsdk:"domain"`
	Purpose            types.String `tfsdk:"purpose"`
	VerificationMethod types.String `tfsdk:"verification_method"`
	Metadata           types.Map    `tfsdk:"metadata"`
	Status             types.String `tfsdk:"status"`
	DNSChallenge       types.String `tfsdk:"dns_challenge"`
	DNSRecordName      types.String `tfsdk:"dns_record_name"`
	VerifiedAt         types.String `tfsdk:"verified_at"`
}

// API request/response structs
type domainVerificationCreateRequest struct {
	Domain             string            `json:"domain"`
	Purpose            string            `json:"purpose"`
	VerificationMethod string            `json:"verification_method,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
}

type domainVerificationUpdateRequest struct {
	Metadata *map[string]string `json:"metadata,omitempty"`
}

type domainVerificationResponse struct {
	ID                 string            `json:"id"`
	Domain             string            `json:"domain"`
	Purpose            string            `json:"purpose"`
	VerificationMethod string            `json:"verification_method"`
	Metadata           map[string]string `json:"metadata"`
	Status             string            `json:"status"`
	DNSChallenge       string            `json:"dns_challenge"`
	DNSRecordName      string            `json:"dns_record_name"`
	VerifiedAt         *string           `json:"verified_at,omitempty"`
}

func (r *DomainVerificationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_verification"
}

func (r *DomainVerificationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		MarkdownDescription: `
Manages a domain verification request in WAYSCloud.

Domain verification proves ownership of a domain for use with WAYSCloud services like email, DKIM, DMARC, and SPF.

~> **Note:** This resource creates the verification request and returns the DNS challenge. You must create the required DNS record separately (e.g., using ` + "`wayscloud_dns_record`" + `). Verification is performed automatically by the WAYSCloud background worker.

## Example Usage

` + "```hcl" + `
resource "wayscloud_domain_verification" "email" {
  domain             = "example.com"
  purpose            = "email"
  verification_method = "dns_txt"
}

# Create the verification DNS record
resource "wayscloud_dns_record" "verification" {
  zone_name = "example.com"
  name      = wayscloud_domain_verification.email.dns_record_name
  type      = "TXT"
  value     = wayscloud_domain_verification.email.dns_challenge
  ttl       = 300
}

output "verification_status" {
  value = wayscloud_domain_verification.email.status
}
` + "```" + `

## Import

Domain verifications can be imported using the verification UUID:

` + "```bash" + `
terraform import wayscloud_domain_verification.email 550e8400-e29b-41d4-a716-446655440000
` + "```" + `

## Authentication

This resource requires a **Personal Access Token (PAT)** instead of an API key:

` + "```bash" + `
export WAYSCLOUD_API_KEY="wayscloud_pat_xxx..."
` + "```" + `

If you also use API key resources, use provider aliases:

` + "```hcl" + `
provider "wayscloud" {
  alias   = "pat"
  api_key = "wayscloud_pat_xxx..."
}

resource "wayscloud_domain_verification" "email" {
  provider = wayscloud.pat
  domain   = "example.com"
  purpose  = "email"
}
` + "```" + `
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the verification request (UUID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Domain to verify (e.g., `example.com`).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"purpose": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Verification purpose: `email`, `dkim`, `dmarc`, `spf`, `general`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"verification_method": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("dns_txt"),
				MarkdownDescription: "Verification method: `dns_txt` (default) or `dns_cname`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"metadata": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Key-value metadata for the verification request.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Verification status: `pending`, `verified`, `failed`, `revoked`.",
			},
			"dns_challenge": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "DNS challenge value to set as TXT or CNAME record.",
			},
			"dns_record_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "DNS record name for the challenge (e.g., `_wayscloud-verify.example.com`).",
			},
			"verified_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the domain was verified (ISO 8601). Null until verified.",
			},
		},
	}
}

func (r *DomainVerificationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DomainVerificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DomainVerificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := domainVerificationCreateRequest{
		Domain:             data.Domain.ValueString(),
		Purpose:            data.Purpose.ValueString(),
		VerificationMethod: data.VerificationMethod.ValueString(),
	}

	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		metadata := make(map[string]string)
		resp.Diagnostics.Append(data.Metadata.ElementsAs(ctx, &metadata, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Metadata = metadata
	}

	tflog.Debug(ctx, "Creating domain verification", map[string]interface{}{
		"domain":  createReq.Domain,
		"purpose": createReq.Purpose,
	})

	respBody, err := r.client.Post(ctx, "/v1/domain-verification/domains", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create domain verification: %s", err))
		return
	}

	var verification domainVerificationResponse
	if err := json.Unmarshal(respBody, &verification); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	r.mapResponseToState(ctx, &data, &verification)

	tflog.Trace(ctx, "Created domain verification", map[string]interface{}{
		"id":     verification.ID,
		"domain": verification.Domain,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainVerificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DomainVerificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	verificationID := data.ID.ValueString()

	tflog.Debug(ctx, "Reading domain verification", map[string]interface{}{
		"id": verificationID,
	})

	respBody, err := r.client.Get(ctx, fmt.Sprintf("/v1/domain-verification/domains/%s", verificationID))
	if err != nil {
		if is404(err) {
			tflog.Debug(ctx, "Domain verification not found, removing from state", map[string]interface{}{
				"id": verificationID,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read domain verification: %s", err))
		return
	}

	var verification domainVerificationResponse
	if err := json.Unmarshal(respBody, &verification); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	r.mapResponseToState(ctx, &data, &verification)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainVerificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DomainVerificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state DomainVerificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	verificationID := state.ID.ValueString()

	// Only metadata is updatable
	if data.Metadata.Equal(state.Metadata) {
		tflog.Debug(ctx, "No changes detected for domain verification", map[string]interface{}{
			"id": verificationID,
		})
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		return
	}

	updateReq := domainVerificationUpdateRequest{}
	metadata := make(map[string]string)
	if !data.Metadata.IsNull() && !data.Metadata.IsUnknown() {
		resp.Diagnostics.Append(data.Metadata.ElementsAs(ctx, &metadata, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	updateReq.Metadata = &metadata

	tflog.Debug(ctx, "Updating domain verification", map[string]interface{}{
		"id": verificationID,
	})

	respBody, err := r.client.Patch(ctx, fmt.Sprintf("/v1/domain-verification/domains/%s", verificationID), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update domain verification: %s", err))
		return
	}

	var verification domainVerificationResponse
	if err := json.Unmarshal(respBody, &verification); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	r.mapResponseToState(ctx, &data, &verification)

	tflog.Trace(ctx, "Updated domain verification", map[string]interface{}{
		"id": verification.ID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DomainVerificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DomainVerificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	verificationID := data.ID.ValueString()

	tflog.Debug(ctx, "Deleting domain verification", map[string]interface{}{
		"id": verificationID,
	})

	_, err := r.client.Delete(ctx, fmt.Sprintf("/v1/domain-verification/domains/%s", verificationID))
	if err != nil {
		if is404(err) {
			tflog.Debug(ctx, "Domain verification already deleted", map[string]interface{}{
				"id": verificationID,
			})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete domain verification: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted domain verification", map[string]interface{}{
		"id": verificationID,
	})
}

func (r *DomainVerificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *DomainVerificationResource) mapResponseToState(ctx context.Context, data *DomainVerificationResourceModel, v *domainVerificationResponse) {
	data.ID = types.StringValue(v.ID)
	data.Domain = types.StringValue(v.Domain)
	data.Purpose = types.StringValue(v.Purpose)
	data.VerificationMethod = types.StringValue(strings.ToLower(v.VerificationMethod))
	data.Status = types.StringValue(v.Status)
	data.DNSChallenge = types.StringValue(v.DNSChallenge)
	data.DNSRecordName = types.StringValue(v.DNSRecordName)

	if v.VerifiedAt != nil {
		data.VerifiedAt = types.StringValue(*v.VerifiedAt)
	} else {
		data.VerifiedAt = types.StringNull()
	}

	if len(v.Metadata) > 0 {
		metadataMap, diags := types.MapValueFrom(ctx, types.StringType, v.Metadata)
		if diags.HasError() {
			data.Metadata = types.MapNull(types.StringType)
		} else {
			data.Metadata = metadataMap
		}
	} else if !data.Metadata.IsNull() {
		// Keep existing metadata state if API returns empty
	} else {
		data.Metadata = types.MapNull(types.StringType)
	}
}
