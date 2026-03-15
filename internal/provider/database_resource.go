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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/wayscloudas/terraform-provider-wayscloud/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DatabaseResource{}
var _ resource.ResourceWithImportState = &DatabaseResource{}

func NewDatabaseResource() resource.Resource {
	return &DatabaseResource{}
}

// DatabaseResource defines the resource implementation.
type DatabaseResource struct {
	client *client.Client
}

// DatabaseResourceModel describes the resource data model.
type DatabaseResourceModel struct {
	Name             types.String `tfsdk:"name"`
	TechnicalName    types.String `tfsdk:"technical_name"`
	Type             types.String `tfsdk:"type"`
	Tier             types.String `tfsdk:"tier"`
	Description      types.String `tfsdk:"description"`
	Host             types.String `tfsdk:"host"`
	Port             types.Int64  `tfsdk:"port"`
	Username         types.String `tfsdk:"username"`
	Password         types.String `tfsdk:"password"`
	ConnectionString types.String `tfsdk:"connection_string"`
	Status           types.String `tfsdk:"status"`
	SizeMB           types.Float64 `tfsdk:"size_mb"`
	CreatedAt        types.String `tfsdk:"created_at"`
}

// API request/response structs
type databaseCreateRequest struct {
	DBName      string `json:"db_name"`
	DBType      string `json:"db_type"`
	Description string `json:"description,omitempty"`
	Tier        string `json:"tier,omitempty"`
}

type databaseCreateResponse struct {
	OK               bool   `json:"ok"`
	Created          bool   `json:"created"`
	FriendlyName     string `json:"friendly_name"`
	TechnicalName    string `json:"technical_name"`
	DBType           string `json:"db_type"`
	Host             string `json:"host"`
	Port             int64  `json:"port"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	ConnectionString string `json:"connection_string"`
}

type databaseInfoResponse struct {
	DBName        string   `json:"db_name"`
	FriendlyName  *string  `json:"friendly_name,omitempty"`
	TechnicalName *string  `json:"technical_name,omitempty"`
	DBType        string   `json:"db_type"`
	Host          string   `json:"host"`
	Port          int64    `json:"port"`
	Status        string   `json:"status"`
	SizeMB        *float64 `json:"size_mb,omitempty"`
	CreatedAt     *string  `json:"created_at,omitempty"`
	Description   *string  `json:"description,omitempty"`
}

func (r *DatabaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *DatabaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version: 0,
		MarkdownDescription: `
Manages a PostgreSQL or MariaDB database instance in WAYSCloud.

WAYSCloud DBaaS provides fully managed database instances with automatic backups,
monitoring, and optional encryption at rest.

~> **Note:** This resource requires a Personal Access Token (PAT) with ` + "`database:read`" + ` and ` + "`database:write`" + ` scopes.
Use ` + "`wayscloud_pat_xxx...`" + ` instead of ` + "`wayscloud_api_xxx...`" + `.

## Example Usage

### PostgreSQL Database

` + "```hcl" + `
resource "wayscloud_database" "app" {
  name        = "myapp-prod"
  type        = "postgresql"
  tier        = "standard"
  description = "Production application database"
}

output "db_host" {
  value = wayscloud_database.app.host
}

output "db_connection" {
  value     = wayscloud_database.app.connection_string
  sensitive = true
}
` + "```" + `

### MariaDB with Encryption

` + "```hcl" + `
resource "wayscloud_database" "secure" {
  name        = "secure-data"
  type        = "mariadb"
  tier        = "encrypted"
  description = "Encrypted database for sensitive data"
}
` + "```" + `

## Import

Databases can be imported using the format ` + "`type/tier/name`" + `:

` + "```bash" + `
terraform import wayscloud_database.app postgresql/standard/myapp-prod
` + "```" + `

Legacy format ` + "`type/name`" + ` is also supported (assumes ` + "`standard`" + ` tier):

` + "```bash" + `
terraform import wayscloud_database.app postgresql/myapp-prod
` + "```" + `

~> **Note:** Password cannot be retrieved after import. You may need to reset it via the dashboard.
`,

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Friendly name for the database (alphanumeric, hyphens, underscores).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"technical_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Technical database name (auto-generated, used internally).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Database type: `postgresql` or `mariadb`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tier": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("standard"),
				MarkdownDescription: "Database tier: `standard` or `encrypted`. Default: `standard`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Description of the database purpose.",
			},
			"host": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Database host endpoint.",
			},
			"port": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Database port (5432 for PostgreSQL, 3306 for MariaDB).",
			},
			"username": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Database username.",
			},
			"password": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Database password. Only available on initial creation.",
			},
			"connection_string": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Full connection string. Only available on initial creation.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Database status: `running`, `unknown`.",
			},
			"size_mb": schema.Float64Attribute{
				Computed:            true,
				MarkdownDescription: "Database size in megabytes.",
			},
			"created_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Timestamp when the database was created (ISO 8601).",
			},
		},
	}
}

func (r *DatabaseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	// Database resource requires PAT (Personal Access Token), not API Key
	if c.PATToken == "" {
		resp.Diagnostics.AddWarning(
			"PAT Token Recommended for Database Resource",
			"The wayscloud_database resource requires a Personal Access Token (PAT) with "+
				"'database:read' and 'database:write' scopes. You are using an API key which may "+
				"not have the required permissions.\n\n"+
				"To create a PAT:\n"+
				"1. Go to my.wayscloud.services → Account → Personal Access Tokens\n"+
				"2. Create a token with 'database:read' and 'database:write' scopes\n"+
				"3. Set WAYSCLOUD_API_KEY=wayscloud_pat_xxx... (the provider auto-detects the token type)",
		)
	}

	r.client = c
}

func (r *DatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := databaseCreateRequest{
		DBName: data.Name.ValueString(),
		DBType: data.Type.ValueString(),
		Tier:   data.Tier.ValueString(),
	}
	if !data.Description.IsNull() {
		createReq.Description = data.Description.ValueString()
	}

	tflog.Debug(ctx, "Creating database", map[string]interface{}{
		"name": createReq.DBName,
		"type": createReq.DBType,
		"tier": createReq.Tier,
	})

	respBody, err := r.client.Post(ctx, "/v1/databases", createReq)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create database: %s", err))
		return
	}

	var createResp databaseCreateResponse
	if err := json.Unmarshal(respBody, &createResp); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Map response to state
	data.Name = types.StringValue(createResp.FriendlyName)
	data.TechnicalName = types.StringValue(createResp.TechnicalName)
	data.Type = types.StringValue(createResp.DBType)
	data.Host = types.StringValue(createResp.Host)
	data.Port = types.Int64Value(createResp.Port)
	data.Username = types.StringValue(createResp.Username)
	data.Password = types.StringValue(createResp.Password)
	data.ConnectionString = types.StringValue(createResp.ConnectionString)
	data.Status = types.StringValue("running")
	data.SizeMB = types.Float64Value(0)

	tflog.Trace(ctx, "Created database", map[string]interface{}{
		"name":           createResp.FriendlyName,
		"technical_name": createResp.TechnicalName,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbName := data.Name.ValueString()
	dbType := data.Type.ValueString()

	tflog.Debug(ctx, "Reading database", map[string]interface{}{
		"name": dbName,
		"type": dbType,
	})

	respBody, err := r.client.Get(ctx, fmt.Sprintf("/v1/databases/%s/%s", dbType, dbName))
	if err != nil {
		// Check if database was deleted
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Database not found, removing from state", map[string]interface{}{
				"name": dbName,
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read database: %s", err))
		return
	}

	var dbInfo databaseInfoResponse
	if err := json.Unmarshal(respBody, &dbInfo); err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	// Preserve secrets from state (not returned on read)
	password := data.Password
	connectionString := data.ConnectionString
	username := data.Username

	// Map response to state
	if dbInfo.FriendlyName != nil {
		data.Name = types.StringValue(*dbInfo.FriendlyName)
	}
	if dbInfo.TechnicalName != nil {
		data.TechnicalName = types.StringValue(*dbInfo.TechnicalName)
	}
	data.Type = types.StringValue(dbInfo.DBType)
	data.Host = types.StringValue(dbInfo.Host)
	data.Port = types.Int64Value(dbInfo.Port)
	data.Status = types.StringValue(dbInfo.Status)
	if dbInfo.SizeMB != nil {
		data.SizeMB = types.Float64Value(*dbInfo.SizeMB)
	}
	if dbInfo.CreatedAt != nil {
		data.CreatedAt = types.StringValue(*dbInfo.CreatedAt)
	}
	if dbInfo.Description != nil {
		data.Description = types.StringValue(*dbInfo.Description)
	}

	// Restore preserved values
	data.Password = password
	data.ConnectionString = connectionString
	data.Username = username

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Databases don't support in-place updates for most fields
	// Description could be updated, but API doesn't support it yet
	var data DatabaseResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DatabaseResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	dbName := data.Name.ValueString()
	dbType := data.Type.ValueString()

	tflog.Debug(ctx, "Deleting database", map[string]interface{}{
		"name": dbName,
		"type": dbType,
	})

	_, err := r.client.Delete(ctx, fmt.Sprintf("/v1/databases/%s/%s", dbType, dbName))
	if err != nil {
		// Ignore 404 errors (already deleted)
		if apiErr, ok := err.(*client.APIError); ok && apiErr.StatusCode == 404 {
			tflog.Debug(ctx, "Database already deleted", map[string]interface{}{
				"name": dbName,
			})
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete database: %s", err))
		return
	}

	tflog.Trace(ctx, "Deleted database", map[string]interface{}{
		"name": dbName,
	})
}

func (r *DatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: type/tier/name (e.g., "postgresql/standard/myapp-prod")
	// Also supports legacy format: type/name (e.g., "postgresql/myapp-prod") — assumes "standard" tier
	parts := splitImportID(req.ID)

	var dbType, dbTier, dbName string
	switch len(parts) {
	case 3:
		dbType = parts[0]
		dbTier = parts[1]
		dbName = parts[2]
	case 2:
		// Legacy format: type/name — assume standard tier
		dbType = parts[0]
		dbTier = "standard"
		dbName = parts[1]
	default:
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID must be in format 'type/tier/name' (e.g., 'postgresql/standard/myapp-prod') "+
				"or legacy format 'type/name' (e.g., 'postgresql/myapp-prod'). Got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), dbType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tier"), dbTier)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), dbName)...)
}

// splitImportID splits an import ID by "/"
func splitImportID(id string) []string {
	var parts []string
	var current string
	for _, c := range id {
		if c == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
