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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &S3BucketKeyResource{}
var _ resource.ResourceWithImportState = &S3BucketKeyResource{}

func NewS3BucketKeyResource() resource.Resource {
	return &S3BucketKeyResource{}
}

// S3BucketKeyResource defines the resource implementation.
type S3BucketKeyResource struct {
	client *client.Client
}

// S3BucketKeyResourceModel describes the resource data model.
type S3BucketKeyResourceModel struct {
	ID         types.String `tfsdk:"id"`
	BucketName types.String `tfsdk:"bucket_name"`
	Name       types.String `tfsdk:"name"`
	AccessKey  types.String `tfsdk:"access_key"`
	SecretKey  types.String `tfsdk:"secret_key"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

// API request/response structs
type s3BucketKeyCreateRequest struct {
	Name string `json:"name"`
}

type s3BucketKeyResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	CreatedAt string `json:"created_at"`
}

func (r *S3BucketKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_bucket_key"
}

func (r *S3BucketKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		MarkdownDescription: `
Manages an S3 bucket access key in WAYSCloud.

Access keys provide programmatic access to S3 buckets. Each key has an access key ID and a secret key. The secret key is only available at creation time and cannot be retrieved later.

## Example Usage

` + "```hcl" + `
resource "wayscloud_s3_bucket" "data" {
  name = "my-data-bucket"
}

resource "wayscloud_s3_bucket_key" "app" {
  bucket_name = wayscloud_s3_bucket.data.name
  name        = "app-key"
}

output "access_key" {
  value = wayscloud_s3_bucket_key.app.access_key
}

output "secret_key" {
  value     = wayscloud_s3_bucket_key.app.secret_key
  sensitive = true
}
` + "```" + `

## Import

S3 bucket keys can be imported using the format ` + "`bucket_name/key_id`" + `:

` + "```bash" + `
terraform import wayscloud_s3_bucket_key.app my-data-bucket/550e8400-e29b-41d4-a716-446655440000
` + "```" + `

~> **Note:** The secret key cannot be retrieved after import. The value in state will be empty.
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the bucket key (UUID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bucket_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The S3 bucket name this key belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Human-readable name for the access key.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"access_key": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The access key ID for S3 authentication.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_key": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The secret key for S3 authentication. Only available on initial creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the key was created (ISO 8601).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *S3BucketKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *S3BucketKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data S3BucketKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := s3BucketKeyCreateRequest{
		Name: data.Name.ValueString(),
	}

	bucketName := data.BucketName.ValueString()

	tflog.Debug(ctx, "Creating S3 bucket key", map[string]interface{}{
		"bucket_name": bucketName,
		"name":        createReq.Name,
	})

	respBody, err := r.client.Post(ctx, fmt.Sprintf("/v1/storage/buckets/%s/keys", bucketName), createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create S3 bucket key: %s", err))
		return
	}

	var key s3BucketKeyResponse
	if err := json.Unmarshal(respBody, &key); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Map response to state
	data.ID = types.StringValue(key.ID)
	data.BucketName = types.StringValue(bucketName)
	data.Name = types.StringValue(key.Name)
	data.AccessKey = types.StringValue(key.AccessKey)
	data.SecretKey = types.StringValue(key.SecretKey)
	data.CreatedAt = types.StringValue(key.CreatedAt)

	tflog.Trace(ctx, "Created S3 bucket key", map[string]interface{}{
		"key_id":      key.ID,
		"bucket_name": bucketName,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *S3BucketKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data S3BucketKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucketName := data.BucketName.ValueString()
	keyID := data.ID.ValueString()

	// Preserve secret key from state (API doesn't return it on read)
	secretKey := data.SecretKey

	tflog.Debug(ctx, "Reading S3 bucket key", map[string]interface{}{
		"bucket_name": bucketName,
		"key_id":      keyID,
	})

	// API does not have a single-key GET endpoint, so list all and filter
	respBody, err := r.client.Get(ctx, fmt.Sprintf("/v1/storage/buckets/%s/keys", bucketName))
	if err != nil {
		if is404(err) {
			tflog.Debug(ctx, "S3 bucket not found, removing key from state", map[string]interface{}{
				"key_id": keyID,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read S3 bucket keys: %s", err))
		return
	}

	// API may return a bare array or {"keys": [...]}
	var keys []s3BucketKeyResponse
	var listResp struct {
		Keys []s3BucketKeyResponse `json:"keys"`
	}
	if err := json.Unmarshal(respBody, &keys); err != nil {
		// Try object wrapper format
		if err2 := json.Unmarshal(respBody, &listResp); err2 != nil {
			resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err2))
			return
		}
		keys = listResp.Keys
	}

	// Find our key by ID
	var found *s3BucketKeyResponse
	for _, k := range keys {
		if k.ID == keyID {
			found = &k
			break
		}
	}

	if found == nil {
		tflog.Debug(ctx, "S3 bucket key not found, removing from state", map[string]interface{}{
			"key_id": keyID,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Map response to state
	data.ID = types.StringValue(found.ID)
	data.BucketName = types.StringValue(bucketName)
	data.Name = types.StringValue(found.Name)
	data.AccessKey = types.StringValue(found.AccessKey)
	data.CreatedAt = types.StringValue(found.CreatedAt)

	// Restore secret key from state (not returned by API on read)
	data.SecretKey = secretKey

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *S3BucketKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All fields are immutable (RequiresReplace), so Update should never be called.
	// If it is, just pass through the plan state.
	var data S3BucketKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *S3BucketKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data S3BucketKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucketName := data.BucketName.ValueString()
	keyID := data.ID.ValueString()

	tflog.Debug(ctx, "Deleting S3 bucket key", map[string]interface{}{
		"bucket_name": bucketName,
		"key_id":      keyID,
	})

	_, err := r.client.Delete(ctx, fmt.Sprintf("/v1/storage/buckets/%s/keys/%s", bucketName, keyID))
	if err != nil {
		// Ignore 404 errors (already deleted)
		if is404(err) {
			tflog.Debug(ctx, "S3 bucket key already deleted", map[string]interface{}{
				"key_id": keyID,
			})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete S3 bucket key: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted S3 bucket key", map[string]interface{}{
		"key_id": keyID,
	})
}

func (r *S3BucketKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: bucket_name/key_id
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: bucket_name/key_id, got: %s", req.ID),
		)
		return
	}

	bucketName := parts[0]
	keyID := parts[1]

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("bucket_name"), bucketName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), keyID)...)
}
