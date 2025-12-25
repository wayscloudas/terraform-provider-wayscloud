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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/wayscloud/terraform-provider-wayscloud/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &RedisInstanceResource{}
var _ resource.ResourceWithImportState = &RedisInstanceResource{}

func NewRedisInstanceResource() resource.Resource {
	return &RedisInstanceResource{}
}

// RedisInstanceResource defines the resource implementation.
type RedisInstanceResource struct {
	client *client.Client
}

// RedisInstanceResourceModel describes the resource data model.
type RedisInstanceResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Region       types.String `tfsdk:"region"`
	Plan         types.String `tfsdk:"plan"`
	Persistence  types.Bool   `tfsdk:"persistence"`
	Status       types.String `tfsdk:"status"`
	Endpoint     types.String `tfsdk:"endpoint"`
	Port         types.Int64  `tfsdk:"port"`
	MemoryMB     types.Int64  `tfsdk:"memory_mb"`
	RedisVersion types.String `tfsdk:"redis_version"`
	Password     types.String `tfsdk:"password"`
	CreatedAt    types.String `tfsdk:"created_at"`
}

// API request/response structs
type redisCreateRequest struct {
	Name        string   `json:"name"`
	Region      string   `json:"region"`
	Plan        string   `json:"plan"`
	Persistence bool     `json:"persistence"`
	AllowCIDRs  []string `json:"allow_cidrs,omitempty"`
}

type redisInstanceResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Region       string  `json:"region"`
	Plan         string  `json:"plan"`
	Status       string  `json:"status"`
	Endpoint     *string `json:"endpoint,omitempty"`
	Port         *int64  `json:"port,omitempty"`
	MemoryMB     int64   `json:"memory_mb"`
	Persistence  bool    `json:"persistence"`
	RedisVersion string  `json:"redis_version"`
	Password     *string `json:"password,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

func (r *RedisInstanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redis_instance"
}

func (r *RedisInstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `
Manages a Redis instance in WAYSCloud.

WAYSCloud Redis provides fully managed Redis instances with automatic failover, backups, and monitoring.

## Example Usage

### Basic Redis Instance

` + "```hcl" + `
resource "wayscloud_redis_instance" "cache" {
  name   = "my-cache"
  region = "no"
  plan   = "redis-starter"
}

output "redis_endpoint" {
  value = wayscloud_redis_instance.cache.endpoint
}

output "redis_password" {
  value     = wayscloud_redis_instance.cache.password
  sensitive = true
}
` + "```" + `

### Redis with Persistence

` + "```hcl" + `
resource "wayscloud_redis_instance" "sessions" {
  name        = "session-store"
  region      = "no"
  plan        = "redis-standard"
  persistence = true
}
` + "```" + `

## Import

Redis instances can be imported using the instance ID:

` + "```bash" + `
terraform import wayscloud_redis_instance.cache 550e8400-e29b-41d4-a716-446655440000
` + "```" + `

~> **Note:** The password cannot be retrieved after import. You may need to rotate it.
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the Redis instance (UUID).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Friendly name for the Redis instance (3-50 characters).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("no"),
				MarkdownDescription: "Region code where the instance will be deployed. Default: `no` (Norway).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"plan": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("redis-starter"),
				MarkdownDescription: "Plan ID. Available plans: `redis-starter`, `redis-standard`, `redis-professional`, `redis-business`, `redis-enterprise`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"persistence": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				MarkdownDescription: "Enable data persistence (RDB snapshots). Default: `true`.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Instance status: `creating`, `running`, `stopped`, `restarting`, `error`, `deleted`.",
			},
			"endpoint": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Redis endpoint hostname.",
			},
			"port": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Redis port number.",
			},
			"memory_mb": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Memory allocation in megabytes.",
			},
			"redis_version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Redis server version.",
			},
			"password": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Redis authentication password. Only available on initial creation.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the instance was created (ISO 8601).",
			},
		},
	}
}

func (r *RedisInstanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RedisInstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RedisInstanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := redisCreateRequest{
		Name:        data.Name.ValueString(),
		Region:      data.Region.ValueString(),
		Plan:        data.Plan.ValueString(),
		Persistence: data.Persistence.ValueBool(),
	}

	tflog.Debug(ctx, "Creating Redis instance", map[string]interface{}{
		"name":   createReq.Name,
		"region": createReq.Region,
		"plan":   createReq.Plan,
	})

	respBody, err := r.client.Post(ctx, "/v1/redis/instances", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create Redis instance: %s", err))
		return
	}

	var instance redisInstanceResponse
	if err := json.Unmarshal(respBody, &instance); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Map response to state (initial state - password is available)
	r.mapResponseToState(&data, &instance)

	tflog.Trace(ctx, "Created Redis instance", map[string]interface{}{
		"id":   instance.ID,
		"name": instance.Name,
	})

	// Save initial state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Wait for instance to be ready
	if instance.Status == "creating" {
		tflog.Debug(ctx, "Waiting for Redis instance to be ready", map[string]interface{}{
			"id": instance.ID,
		})

		readyInstance, err := r.waitForReady(ctx, instance.ID)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Instance Still Creating",
				fmt.Sprintf("Redis instance created but not yet ready. Status: %s. Run terraform refresh to update.", instance.Status),
			)
			return
		}

		// Update with final state (but keep password from create response)
		password := data.Password
		r.mapResponseToState(&data, readyInstance)
		data.Password = password // Preserve password

		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

func (r *RedisInstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RedisInstanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceID := data.ID.ValueString()

	tflog.Debug(ctx, "Reading Redis instance", map[string]interface{}{
		"id": instanceID,
	})

	respBody, err := r.client.Get(ctx, fmt.Sprintf("/v1/redis/instances/%s", instanceID))
	if err != nil {
		// Check if instance was deleted
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Redis instance not found, removing from state", map[string]interface{}{
				"id": instanceID,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read Redis instance: %s", err))
		return
	}

	var instance redisInstanceResponse
	if err := json.Unmarshal(respBody, &instance); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Preserve password from state (not returned on read)
	password := data.Password
	r.mapResponseToState(&data, &instance)
	data.Password = password

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RedisInstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Redis instances don't support updates (all fields require replacement)
	// Just read the current state
	var data RedisInstanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RedisInstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RedisInstanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	instanceID := data.ID.ValueString()

	tflog.Debug(ctx, "Deleting Redis instance", map[string]interface{}{
		"id": instanceID,
	})

	_, err := r.client.Delete(ctx, fmt.Sprintf("/v1/redis/instances/%s", instanceID))
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Redis instance already deleted", map[string]interface{}{
				"id": instanceID,
			})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete Redis instance: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted Redis instance", map[string]interface{}{
		"id": instanceID,
	})
}

func (r *RedisInstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapResponseToState maps the API response to the Terraform state model
func (r *RedisInstanceResource) mapResponseToState(data *RedisInstanceResourceModel, instance *redisInstanceResponse) {
	data.ID = types.StringValue(instance.ID)
	data.Name = types.StringValue(instance.Name)
	data.Region = types.StringValue(instance.Region)
	data.Plan = types.StringValue(instance.Plan)
	data.Status = types.StringValue(instance.Status)
	data.MemoryMB = types.Int64Value(instance.MemoryMB)
	data.Persistence = types.BoolValue(instance.Persistence)
	data.RedisVersion = types.StringValue(instance.RedisVersion)
	data.CreatedAt = types.StringValue(instance.CreatedAt)

	// Handle optional fields
	if instance.Endpoint != nil {
		data.Endpoint = types.StringValue(*instance.Endpoint)
	} else {
		data.Endpoint = types.StringNull()
	}
	if instance.Port != nil {
		data.Port = types.Int64Value(*instance.Port)
	} else {
		data.Port = types.Int64Null()
	}
	if instance.Password != nil {
		data.Password = types.StringValue(*instance.Password)
	}
	// Note: Don't set password to null if not in response - preserve existing value
}

// waitForReady polls the instance until it's ready or timeout
func (r *RedisInstanceResource) waitForReady(ctx context.Context, instanceID string) (*redisInstanceResponse, error) {
	timeout := 5 * time.Minute
	pollInterval := 10 * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		respBody, err := r.client.Get(ctx, fmt.Sprintf("/v1/redis/instances/%s", instanceID))
		if err != nil {
			return nil, err
		}

		var instance redisInstanceResponse
		if err := json.Unmarshal(respBody, &instance); err != nil {
			return nil, err
		}

		tflog.Debug(ctx, "Polling Redis instance status", map[string]interface{}{
			"id":     instanceID,
			"status": instance.Status,
		})

		switch instance.Status {
		case "running":
			return &instance, nil
		case "error", "deleted":
			return nil, fmt.Errorf("instance entered %s state", instance.Status)
		case "creating", "restarting":
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

	return nil, fmt.Errorf("timeout waiting for instance to be ready")
}
