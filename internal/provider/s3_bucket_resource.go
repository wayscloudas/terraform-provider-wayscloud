// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"strings"
	"time"
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
var _ resource.Resource = &S3BucketResource{}
var _ resource.ResourceWithImportState = &S3BucketResource{}

func NewS3BucketResource() resource.Resource {
	return &S3BucketResource{}
}

// S3BucketResource defines the resource implementation.
type S3BucketResource struct {
	client *client.Client
}

// S3BucketResourceModel describes the resource data model.
type S3BucketResourceModel struct {
	BucketName     types.String  `tfsdk:"bucket_name"`
	Tier           types.String  `tfsdk:"tier"`
	Region         types.String  `tfsdk:"region"`
	Endpoint       types.String  `tfsdk:"endpoint"`
	AccessKey      types.String  `tfsdk:"access_key"`
	SecretKey      types.String  `tfsdk:"secret_key"`
	IsActive       types.Bool    `tfsdk:"is_active"`
	TotalStorageGB types.Float64 `tfsdk:"total_storage_gb"`
	TotalObjects   types.Int64   `tfsdk:"total_objects"`
	CreatedAt      types.String  `tfsdk:"created_at"`
}

// API request/response structs
type s3BucketCreateRequest struct {
	BucketName string `json:"bucket_name"`
	Tier       string `json:"tier"`
}

type s3BucketCreateResponse struct {
	Success    bool   `json:"success"`
	BucketName string `json:"bucket_name"`
	Message    string `json:"message"`
	Endpoint   string `json:"endpoint"`
	Region     string `json:"region"`
	Tier       string `json:"tier"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
}

type s3BucketResponse struct {
	BucketName     string   `json:"bucket_name"`
	StorageTier    string   `json:"storage_tier"`
	Region         string   `json:"region"`
	CreatedAt      *string  `json:"created_at,omitempty"`
	IsActive       bool     `json:"is_active"`
	IsPublic       bool     `json:"is_public"`
	TotalStorageGB *float64 `json:"total_storage_gb,omitempty"`
	TotalObjects   *int64   `json:"total_objects,omitempty"`
}

func (r *S3BucketResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_bucket"
}

func (r *S3BucketResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		MarkdownDescription: `
Manages an S3-compatible storage bucket in WAYSCloud.

WAYSCloud Storage provides S3-compatible object storage with support for standard and enterprise tiers.

## Example Usage

### Standard Bucket

` + "```hcl" + `
resource "wayscloud_s3_bucket" "uploads" {
  bucket_name = "my-app-uploads"
  tier        = "standard"
}

output "s3_endpoint" {
  value = wayscloud_s3_bucket.uploads.endpoint
}

output "s3_access_key" {
  value = wayscloud_s3_bucket.uploads.access_key
}

output "s3_secret_key" {
  value     = wayscloud_s3_bucket.uploads.secret_key
  sensitive = true
}
` + "```" + `

### Enterprise Bucket (High Performance)

` + "```hcl" + `
resource "wayscloud_s3_bucket" "critical_data" {
  bucket_name = "production-data"
  tier        = "enterprise"
}
` + "```" + `

## S3 Client Configuration

Use the outputs to configure your S3 client:

` + "```hcl" + `
# Example for AWS SDK/CLI configuration
output "aws_config" {
  value = <<-EOT
    [profile wayscloud]
    aws_access_key_id = ${wayscloud_s3_bucket.uploads.access_key}
    aws_secret_access_key = ${wayscloud_s3_bucket.uploads.secret_key}
    endpoint_url = ${wayscloud_s3_bucket.uploads.endpoint}
    region = eu-north-1
  EOT
  sensitive = true
}
` + "```" + `

## Import

S3 buckets can be imported using the bucket name:

` + "```bash" + `
terraform import wayscloud_s3_bucket.uploads my-app-uploads
` + "```" + `

~> **Note:** Secret key cannot be retrieved after import. Create a new access key if needed.
`,

		Attributes: map[string]schema.Attribute{
			"bucket_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Globally unique bucket name. Must be 3-63 characters, lowercase letters, numbers, and hyphens only. Cannot start or end with hyphen.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tier": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("standard"),
				MarkdownDescription: "Storage tier: `standard` (general use) or `enterprise` (dedicated infrastructure, enhanced performance). Default: `standard`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Storage region (currently `no-oslo-1`).",
			},
			"endpoint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "S3 API endpoint URL.",
			},
			"access_key": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "S3 access key ID.",
			},
			"secret_key": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "S3 secret access key. Only available on initial creation.",
			},
			"is_active": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the bucket is active.",
			},
			"total_storage_gb": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "Total storage used in gigabytes.",
			},
			"total_objects": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Total number of objects in the bucket.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the bucket was created (ISO 8601).",
			},
		},
	}
}

func (r *S3BucketResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *S3BucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data S3BucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := s3BucketCreateRequest{
		BucketName: data.BucketName.ValueString(),
		Tier:       data.Tier.ValueString(),
	}

	tflog.Debug(ctx, "Creating S3 bucket", map[string]interface{}{
		"bucket_name": createReq.BucketName,
		"tier":        createReq.Tier,
	})

	respBody, err := r.client.Post(ctx, "/v1/storage/buckets", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create S3 bucket: %s", err))
		return
	}

	var createResp s3BucketCreateResponse
	if err := json.Unmarshal(respBody, &createResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Map response to state
	data.BucketName = types.StringValue(createResp.BucketName)
	data.Tier = types.StringValue(createResp.Tier)
	data.Region = types.StringValue(createResp.Region)
	data.Endpoint = types.StringValue(createResp.Endpoint)
	data.AccessKey = types.StringValue(createResp.AccessKey)
	data.SecretKey = types.StringValue(createResp.SecretKey)
	data.IsActive = types.BoolValue(true)
	data.TotalStorageGB = types.Float64Value(0)
	data.TotalObjects = types.Int64Value(0)
	data.CreatedAt = types.StringValue(time.Now().UTC().Format(time.RFC3339))

	tflog.Trace(ctx, "Created S3 bucket", map[string]interface{}{
		"bucket_name": createResp.BucketName,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *S3BucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data S3BucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucketName := data.BucketName.ValueString()

	tflog.Debug(ctx, "Reading S3 bucket", map[string]interface{}{
		"bucket_name": bucketName,
	})

	respBody, err := r.client.Get(ctx, fmt.Sprintf("/v1/storage/buckets/%s", bucketName))
	if err != nil {
		// Check if bucket was deleted
		if is404(err) {
			tflog.Debug(ctx, "S3 bucket not found, removing from state", map[string]interface{}{
				"bucket_name": bucketName,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read S3 bucket: %s", err))
		return
	}

	var bucket s3BucketResponse
	if err := json.Unmarshal(respBody, &bucket); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Preserve values from state (not returned on read)
	accessKey := data.AccessKey
	secretKey := data.SecretKey
	endpoint := data.Endpoint

	// Map response to state
	data.BucketName = types.StringValue(bucket.BucketName)
	data.Tier = types.StringValue(strings.ToLower(bucket.StorageTier))
	data.Region = types.StringValue(bucket.Region)
	data.IsActive = types.BoolValue(bucket.IsActive)
	// Preserve endpoint from state (not returned by API on read)
	data.Endpoint = endpoint

	if bucket.TotalStorageGB != nil {
		data.TotalStorageGB = types.Float64Value(*bucket.TotalStorageGB)
	} else {
		data.TotalStorageGB = types.Float64Value(0)
	}
	if bucket.TotalObjects != nil {
		data.TotalObjects = types.Int64Value(*bucket.TotalObjects)
	} else {
		data.TotalObjects = types.Int64Value(0)
	}
	if bucket.CreatedAt != nil {
		data.CreatedAt = types.StringValue(*bucket.CreatedAt)
	}

	// Restore preserved values
	data.AccessKey = accessKey
	data.SecretKey = secretKey

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *S3BucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// S3 buckets don't support updates (all fields require replacement)
	var data S3BucketResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *S3BucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data S3BucketResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucketName := data.BucketName.ValueString()

	tflog.Debug(ctx, "Deleting S3 bucket", map[string]interface{}{
		"bucket_name": bucketName,
	})

	_, err := r.client.Delete(ctx, fmt.Sprintf("/v1/storage/buckets/%s", bucketName))
	if err != nil {
		// Ignore 404 errors (already deleted)
		if is404(err) {
			tflog.Debug(ctx, "S3 bucket already deleted", map[string]interface{}{
				"bucket_name": bucketName,
			})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete S3 bucket: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted S3 bucket", map[string]interface{}{
		"bucket_name": bucketName,
	})
}

func (r *S3BucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by bucket name
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("bucket_name"), req.ID)...)
	// Set default endpoint for imported buckets (will be preserved on subsequent reads)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("endpoint"), "https://storage.wayscloud.services")...)
}
