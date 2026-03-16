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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &AppResource{}
var _ resource.ResourceWithImportState = &AppResource{}

func NewAppResource() resource.Resource {
	return &AppResource{}
}

// AppResource defines the resource implementation.
type AppResource struct {
	client *client.Client
}

// AppResourceModel describes the resource data model.
type AppResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Slug                 types.String `tfsdk:"slug"`
	ShortID              types.String `tfsdk:"short_id"`
	Region               types.String `tfsdk:"region"`
	Plan                 types.String `tfsdk:"plan"`
	PlanName             types.String `tfsdk:"plan_name"`
	Status               types.String `tfsdk:"status"`
	Port                 types.Int64  `tfsdk:"port"`
	HealthCheckPath      types.String `tfsdk:"health_check_path"`
	EnvVars              types.Map    `tfsdk:"env_vars"`
	MinInstances         types.Int64  `tfsdk:"min_instances"`
	MaxInstances         types.Int64  `tfsdk:"max_instances"`
	ScaleToZeroEnabled   types.Bool   `tfsdk:"scale_to_zero_enabled"`
	IdleTimeoutMinutes   types.Int64  `tfsdk:"idle_timeout_minutes"`
	DefaultURL           types.String `tfsdk:"default_url"`
	ActiveImage          types.String `tfsdk:"active_image"`
	ActiveRevisionID     types.String `tfsdk:"active_revision_id"`
	ActiveDeploymentID   types.String `tfsdk:"active_deployment_id"`
	IsScaledToZero       types.Bool   `tfsdk:"is_scaled_to_zero"`
	CreatedAt            types.String `tfsdk:"created_at"`
	LastDeployedAt       types.String `tfsdk:"last_deployed_at"`
}

// API request/response structs
type appCreateRequest struct {
	Name            string            `json:"name"`
	Slug            string            `json:"slug,omitempty"`
	Region          string            `json:"region,omitempty"`
	Plan            string            `json:"plan,omitempty"`
	Port            int64             `json:"port,omitempty"`
	HealthCheckPath string            `json:"health_check_path,omitempty"`
	EnvVars         map[string]string `json:"env_vars,omitempty"`
}

type appUpdateRequest struct {
	Name                 *string           `json:"name,omitempty"`
	Port                 *int64            `json:"port,omitempty"`
	HealthCheckPath      *string           `json:"health_check_path,omitempty"`
	MinInstances         *int64            `json:"min_instances,omitempty"`
	MaxInstances         *int64            `json:"max_instances,omitempty"`
	EnvVars              map[string]string `json:"env_vars,omitempty"`
	ScaleToZeroEnabled   *bool             `json:"scale_to_zero_enabled,omitempty"`
	IdleTimeoutMinutes   *int64            `json:"idle_timeout_minutes,omitempty"`
}

type appResponse struct {
	ID                   string            `json:"id"`
	Name                 string            `json:"name"`
	Slug                 string            `json:"slug"`
	ShortID              string            `json:"short_id"`
	Region               string            `json:"region"`
	PlanID               string            `json:"plan_id"`
	PlanName             string            `json:"plan_name"`
	Status               string            `json:"status"`
	Port                 int64             `json:"port"`
	HealthCheckPath      string            `json:"health_check_path"`
	MinInstances         int64             `json:"min_instances"`
	MaxInstances         int64             `json:"max_instances"`
	ScaleToZeroEnabled   bool              `json:"scale_to_zero_enabled"`
	IdleTimeoutMinutes   int64             `json:"idle_timeout_minutes"`
	DefaultURL           *string           `json:"default_url,omitempty"`
	ActiveImage          *string           `json:"active_image,omitempty"`
	ActiveRevisionID     *string           `json:"active_revision_id,omitempty"`
	ActiveDeploymentID   *string           `json:"active_deployment_id,omitempty"`
	IsScaledToZero       bool              `json:"is_scaled_to_zero"`
	CreatedAt            string            `json:"created_at"`
	LastDeployedAt       *string           `json:"last_deployed_at,omitempty"`
}

func (r *AppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

func (r *AppResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		MarkdownDescription: `
Manages a container app in WAYSCloud App Platform.

WAYSCloud App Platform provides fully managed container hosting with automatic
scaling, TLS, and zero-downtime deployments.

## Example Usage

### Basic App

` + "```hcl" + `
resource "wayscloud_app" "api" {
  name   = "My API"
  region = "no"
  plan   = "app-basic"

  env_vars = {
    NODE_ENV     = "production"
    DATABASE_URL = wayscloud_database.app.connection_string
  }
}

output "app_url" {
  value = wayscloud_app.api.default_url
}
` + "```" + `

### App with Scale-to-Zero

` + "```hcl" + `
resource "wayscloud_app" "worker" {
  name   = "Background Worker"
  region = "no"
  plan   = "app-basic"

  port              = 3000
  health_check_path = "/healthz"

  scale_to_zero_enabled = true
  idle_timeout_minutes  = 15
  min_instances         = 0
  max_instances         = 3
}
` + "```" + `

## Deployment

After creating an app, use the WAYSCloud CLI or API to deploy:

` + "```bash" + `
# Deploy from Docker Hub
curl -X POST "https://api.wayscloud.services/v1/apps/${app_id}/deploy/image" \
  -H "X-API-Key: $WAYSCLOUD_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"image_uri": "nginx:latest"}'
` + "```" + `

## Import

Apps can be imported using the app ID:

` + "```bash" + `
terraform import wayscloud_app.api app_01ARZ3NDEKTSV4RRFFQ69G5FAV
` + "```" + `
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the app (ULID with prefix).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Display name for the app (3-100 characters).",
			},
			"slug": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "URL slug (auto-generated from name if not provided). Must be lowercase, start with a letter.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"short_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Short identifier for display.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"region": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("no"),
				MarkdownDescription: "Region code. Default: `no` (Norway).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"plan": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("app-basic"),
				MarkdownDescription: "Plan ID. Default: `app-basic`. Available: `app-basic`, `app-standard`, `app-professional`.",
			},
			"plan_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Human-readable plan name.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "App status: `creating`, `building`, `deploying`, `running`, `stopped`, `error`.",
			},
			"port": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(8080),
				MarkdownDescription: "Port the app listens on. Default: `8080`.",
			},
			"health_check_path": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("/health"),
				MarkdownDescription: "Health check endpoint. Default: `/health`.",
			},
			"env_vars": schema.MapAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: "Environment variables as key-value pairs.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"min_instances": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
				MarkdownDescription: "Minimum instances. Set to 0 for scale-to-zero. Default: `0`.",
			},
			"max_instances": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				MarkdownDescription: "Maximum instances (1-10). Default: `1`.",
			},
			"scale_to_zero_enabled": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "Enable scale-to-zero on idle. Default: `false`.",
			},
			"idle_timeout_minutes": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(15),
				MarkdownDescription: "Minutes of inactivity before scaling to zero (1-60). Default: `15`.",
			},
			"default_url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Default URL for the app (e.g., `https://myapp.apps.wayscloud.services`).",
			},
			"active_image": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Currently deployed container image.",
			},
			"active_revision_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Currently active revision ID.",
			},
			"active_deployment_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Currently active deployment ID.",
			},
			"is_scaled_to_zero": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the app is currently scaled to zero.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the app was created (ISO 8601).",
			},
			"last_deployed_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp of last deployment (ISO 8601).",
			},
		},
	}
}

func (r *AppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AppResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := appCreateRequest{
		Name:            data.Name.ValueString(),
		Region:          data.Region.ValueString(),
		Plan:            data.Plan.ValueString(),
		Port:            data.Port.ValueInt64(),
		HealthCheckPath: data.HealthCheckPath.ValueString(),
	}

	if !data.Slug.IsNull() && !data.Slug.IsUnknown() {
		createReq.Slug = data.Slug.ValueString()
	}

	// Extract env vars from map
	if !data.EnvVars.IsNull() {
		envVars := make(map[string]string)
		resp.Diagnostics.Append(data.EnvVars.ElementsAs(ctx, &envVars, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.EnvVars = envVars
	}

	tflog.Debug(ctx, "Creating app", map[string]interface{}{
		"name":   createReq.Name,
		"region": createReq.Region,
		"plan":   createReq.Plan,
	})

	respBody, err := r.client.Post(ctx, "/v1/apps", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create app: %s", err))
		return
	}

	var app appResponse
	if err := json.Unmarshal(respBody, &app); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Map response to state, preserving env_vars from plan (API may not echo them back)
	envVars := data.EnvVars
	r.mapResponseToState(ctx, &data, &app)
	data.EnvVars = envVars

	tflog.Trace(ctx, "Created app", map[string]interface{}{
		"id":   app.ID,
		"name": app.Name,
		"slug": app.Slug,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.ID.ValueString()

	tflog.Debug(ctx, "Reading app", map[string]interface{}{
		"id": appID,
	})

	respBody, err := r.client.Get(ctx, fmt.Sprintf("/v1/apps/%s", appID))
	if err != nil {
		// Check if app was deleted
		if is404(err) {
			tflog.Debug(ctx, "App not found, removing from state", map[string]interface{}{
				"id": appID,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read app: %s", err))
		return
	}

	var app appResponse
	if err := json.Unmarshal(respBody, &app); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Map response to state, preserving env_vars from state (API may not echo them back)
	envVars := data.EnvVars
	r.mapResponseToState(ctx, &data, &app)
	data.EnvVars = envVars

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data AppResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state AppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := state.ID.ValueString()

	// Build update request with only changed fields
	updateReq := appUpdateRequest{}

	if data.Name.ValueString() != state.Name.ValueString() {
		name := data.Name.ValueString()
		updateReq.Name = &name
	}
	if data.Port.ValueInt64() != state.Port.ValueInt64() {
		port := data.Port.ValueInt64()
		updateReq.Port = &port
	}
	if data.HealthCheckPath.ValueString() != state.HealthCheckPath.ValueString() {
		hcp := data.HealthCheckPath.ValueString()
		updateReq.HealthCheckPath = &hcp
	}
	if data.MinInstances.ValueInt64() != state.MinInstances.ValueInt64() {
		min := data.MinInstances.ValueInt64()
		updateReq.MinInstances = &min
	}
	if data.MaxInstances.ValueInt64() != state.MaxInstances.ValueInt64() {
		max := data.MaxInstances.ValueInt64()
		updateReq.MaxInstances = &max
	}
	if data.ScaleToZeroEnabled.ValueBool() != state.ScaleToZeroEnabled.ValueBool() {
		stz := data.ScaleToZeroEnabled.ValueBool()
		updateReq.ScaleToZeroEnabled = &stz
	}
	if data.IdleTimeoutMinutes.ValueInt64() != state.IdleTimeoutMinutes.ValueInt64() {
		itm := data.IdleTimeoutMinutes.ValueInt64()
		updateReq.IdleTimeoutMinutes = &itm
	}

	// Handle env vars
	if !data.EnvVars.IsNull() {
		envVars := make(map[string]string)
		resp.Diagnostics.Append(data.EnvVars.ElementsAs(ctx, &envVars, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.EnvVars = envVars
	}

	tflog.Debug(ctx, "Updating app", map[string]interface{}{
		"id": appID,
	})

	respBody, err := r.client.Patch(ctx, fmt.Sprintf("/v1/apps/%s", appID), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update app: %s", err))
		return
	}

	var app appResponse
	if err := json.Unmarshal(respBody, &app); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Map response to state, preserving env_vars from plan (API may not echo them back)
	envVars := data.EnvVars
	r.mapResponseToState(ctx, &data, &app)
	data.EnvVars = envVars

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	appID := data.ID.ValueString()

	tflog.Debug(ctx, "Deleting app", map[string]interface{}{
		"id": appID,
	})

	_, err := r.client.Delete(ctx, fmt.Sprintf("/v1/apps/%s", appID))
	if err != nil {
		// Ignore 404 errors (already deleted)
		if is404(err) {
			tflog.Debug(ctx, "App already deleted", map[string]interface{}{
				"id": appID,
			})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete app: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted app", map[string]interface{}{
		"id": appID,
	})
}

func (r *AppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapResponseToState maps the API response to the Terraform state model
func (r *AppResource) mapResponseToState(ctx context.Context, data *AppResourceModel, app *appResponse) {
	data.ID = types.StringValue(app.ID)
	data.Name = types.StringValue(app.Name)
	data.Slug = types.StringValue(app.Slug)
	data.ShortID = types.StringValue(app.ShortID)
	data.Region = types.StringValue(app.Region)
	data.Plan = types.StringValue(app.PlanID)
	data.PlanName = types.StringValue(app.PlanName)
	data.Status = types.StringValue(app.Status)
	data.Port = types.Int64Value(app.Port)
	data.HealthCheckPath = types.StringValue(app.HealthCheckPath)
	data.MinInstances = types.Int64Value(app.MinInstances)
	data.MaxInstances = types.Int64Value(app.MaxInstances)
	data.ScaleToZeroEnabled = types.BoolValue(app.ScaleToZeroEnabled)
	data.IdleTimeoutMinutes = types.Int64Value(app.IdleTimeoutMinutes)
	data.IsScaledToZero = types.BoolValue(app.IsScaledToZero)
	data.CreatedAt = types.StringValue(app.CreatedAt)

	// Handle optional fields
	if app.DefaultURL != nil {
		data.DefaultURL = types.StringValue(*app.DefaultURL)
	} else {
		data.DefaultURL = types.StringNull()
	}
	if app.ActiveImage != nil {
		data.ActiveImage = types.StringValue(*app.ActiveImage)
	} else {
		data.ActiveImage = types.StringNull()
	}
	if app.ActiveRevisionID != nil {
		data.ActiveRevisionID = types.StringValue(*app.ActiveRevisionID)
	} else {
		data.ActiveRevisionID = types.StringNull()
	}
	if app.ActiveDeploymentID != nil {
		data.ActiveDeploymentID = types.StringValue(*app.ActiveDeploymentID)
	} else {
		data.ActiveDeploymentID = types.StringNull()
	}
	if app.LastDeployedAt != nil {
		data.LastDeployedAt = types.StringValue(*app.LastDeployedAt)
	} else {
		data.LastDeployedAt = types.StringNull()
	}
}

