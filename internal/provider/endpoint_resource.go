package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = endpointResource{}
var _ resource.ResourceWithImportState = endpointResource{}

type endpointResource struct {
	client *resty.Client
}

type endpointResourceJSON struct {
	Host                  string                `json:"host"`
	Id                    string                `json:"id"`
	ProjectID             string                `json:"project_id"`
	BranchID              string                `json:"branch_id"`
	AutoscalingLimitMinCu int64                 `json:"autoscaling_limit_min_cu"`
	AutoscalingLimitMaxCu int64                 `json:"autoscaling_limit_max_cu"`
	RegionID              string                `json:"region_id"`
	Type                  string                `json:"type"`
	CurrentState          string                `json:"current_state"`
	PendingState          string                `json:"pending_state"`
	Settings              *endpointSettingsJSON `json:"settings"`
	PoolerEnabled         bool                  `json:"pooler_enabled"`
	PoolerMode            string                `json:"pooler_mode"`
	Disabled              bool                  `json:"disabled"`
	PasswordlessAccess    bool                  `json:"passwordless_access"`
	LastActive            string                `json:"last_active"`
	CreatedAt             string                `json:"created_at"`
	UpdatedAt             string                `json:"updated_at"`
}
type endpointResourceModel struct {
	Host                  types.String `tfsdk:"host"`
	Id                    types.String `tfsdk:"id"`
	ProjectID             types.String `tfsdk:"project_id"`
	BranchID              types.String `tfsdk:"branch_id"`
	AutoscalingLimitMinCu types.Int64  `tfsdk:"autoscaling_limit_min_cu"`
	AutoscalingLimitMaxCu types.Int64  `tfsdk:"autoscaling_limit_max_cu"`
	RegionID              types.String `tfsdk:"region_id"`
	Type                  types.String `tfsdk:"type"`
	CurrentState          types.String `tfsdk:"current_state"`
	PendingState          types.String `tfsdk:"pending_state"`
	PoolerEnabled         types.Bool   `tfsdk:"pooler_enabled"`
	PoolerMode            types.String `tfsdk:"pooler_mode"`
	Disabled              types.Bool   `tfsdk:"disabled"`
	PasswordlessAccess    types.Bool   `tfsdk:"passwordless_access"`
	LastActive            types.String `tfsdk:"last_active"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
}

func (m *endpointResourceModel) ToEndpointResourceJSON() *endpointResourceJSON {
	return &endpointResourceJSON{
		Host:                  m.Host.ValueString(),
		Id:                    m.Id.ValueString(),
		ProjectID:             m.ProjectID.ValueString(),
		BranchID:              m.BranchID.ValueString(),
		AutoscalingLimitMinCu: m.AutoscalingLimitMinCu.ValueInt64(),
		AutoscalingLimitMaxCu: m.AutoscalingLimitMaxCu.ValueInt64(),
		RegionID:              m.RegionID.ValueString(),
		Type:                  m.Type.ValueString(),
		CurrentState:          m.CurrentState.ValueString(),
		PendingState:          m.PendingState.ValueString(),
		PoolerEnabled:         m.PoolerEnabled.ValueBool(),
		PoolerMode:            m.PoolerMode.ValueString(),
		Disabled:              m.Disabled.ValueBool(),
		PasswordlessAccess:    m.PasswordlessAccess.ValueBool(),
		LastActive:            m.LastActive.ValueString(),
		CreatedAt:             m.CreatedAt.ValueString(),
		UpdatedAt:             m.UpdatedAt.ValueString(),
	}
}

func (m *endpointResourceJSON) ToEndpointResourceModel() *endpointResourceModel {
	return &endpointResourceModel{
		Host:                  types.StringValue(m.Host),
		Id:                    types.StringValue(m.Id),
		ProjectID:             types.StringValue(m.ProjectID),
		BranchID:              types.StringValue(m.BranchID),
		AutoscalingLimitMinCu: types.Int64Value(m.AutoscalingLimitMinCu),
		AutoscalingLimitMaxCu: types.Int64Value(m.AutoscalingLimitMaxCu),
		RegionID:              types.StringValue(m.RegionID),
		Type:                  types.StringValue(m.Type),
		CurrentState:          types.StringValue(m.CurrentState),
		PendingState:          types.StringValue(m.PendingState),
		PoolerEnabled:         types.BoolValue(m.PoolerEnabled),
		PoolerMode:            types.StringValue(m.PoolerMode),
		Disabled:              types.BoolValue(m.Disabled),
		PasswordlessAccess:    types.BoolValue(m.PasswordlessAccess),
		LastActive:            types.StringValue(m.LastActive),
		CreatedAt:             types.StringValue(m.CreatedAt),
		UpdatedAt:             types.StringValue(m.UpdatedAt),
	}
}

func (r endpointResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoint"
}

func (r endpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Neon endpoint resource",
		Attributes:          endpointResourceAttr(),
	}
}

func endpointResourceAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"host": schema.StringAttribute{
			MarkdownDescription: "neon host",
			Computed:            true,
		},
		"id": schema.StringAttribute{
			MarkdownDescription: "endpoint id",
			Computed:            true,
		},
		"project_id": schema.StringAttribute{
			MarkdownDescription: "project id",
			Required:            true,
		},
		"branch_id": schema.StringAttribute{
			MarkdownDescription: "postgres branch",
			Required:            true,
		},
		"autoscaling_limit_min_cu": schema.Int64Attribute{
			MarkdownDescription: "autoscaling limit min",
			Computed:            true,
			Optional:            true,
		},
		"autoscaling_limit_max_cu": schema.Int64Attribute{
			MarkdownDescription: "autoscaling limit max",
			Computed:            true,
			Optional:            true,
		},
		"region_id": schema.StringAttribute{
			MarkdownDescription: "region id",
			Computed:            true,
			Optional:            true,
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "type",
			Required:            true,
			Validators:          []validator.String{stringvalidator.OneOf("read_write", "read_only")},
		},
		"current_state": schema.StringAttribute{
			MarkdownDescription: "current state",
			Computed:            true,
			Validators:          []validator.String{stringvalidator.OneOf("init", "active", "idle")},
		},
		"pending_state": schema.StringAttribute{
			MarkdownDescription: "pending state",
			Computed:            true,
			Validators:          []validator.String{stringvalidator.OneOf("init", "active", "idle")},
		},
		"pooler_enabled": schema.BoolAttribute{
			MarkdownDescription: "pooler enabled",
			Computed:            true,
			Optional:            true,
		},
		"pooler_mode": schema.StringAttribute{
			MarkdownDescription: "pooler mode",
			Computed:            true,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.OneOf("transaction")},
		},
		"disabled": schema.BoolAttribute{
			MarkdownDescription: "disabled",
			Computed:            true,
			Optional:            true,
		},
		"passwordless_access": schema.BoolAttribute{
			MarkdownDescription: "passwordless access",
			Computed:            true,
			Optional:            true,
		},
		//validate dates
		"last_active": schema.StringAttribute{
			MarkdownDescription: "last active",
			Computed:            true,
		},
		"created_at": schema.StringAttribute{
			MarkdownDescription: "created at",
			Computed:            true,
		},
		"updated_at": schema.StringAttribute{
			MarkdownDescription: "updated at",
			Computed:            true,
		},
	}
}

type createEndpoint struct {
	Branch_id                string                `json:"branch_id"`
	Region_id                string                `json:"region_id,omitempty"`
	Type                     string                `json:"type"`
	Settings                 *endpointSettingsJSON `json:"settings,omitempty"`
	Autoscaling_limit_min_cu int64                 `json:"autoscaling_limit_min_cu,omitempty"`
	Autoscaling_limit_max_cu int64                 `json:"autoscaling_limit_max_cu,omitempty"`
	Pooler_enabled           bool                  `json:"pooler_enabled,omitempty"`
	Pooler_mode              string                `json:"pooler_mode,omitempty"`
	Disabled                 bool                  `json:"disabled,omitempty"`
	Passwordless_access      bool                  `json:"passwordless_access,omitempty"`
}

func (r endpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data endpointResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	content := struct {
		Endpoint createEndpoint `json:"endpoint"`
	}{
		Endpoint: createEndpoint{
			Branch_id:                data.BranchID.ValueString(),
			Region_id:                data.RegionID.ValueString(),
			Type:                     data.Type.ValueString(),
			Autoscaling_limit_min_cu: data.AutoscalingLimitMinCu.ValueInt64(),
			Autoscaling_limit_max_cu: data.AutoscalingLimitMaxCu.ValueInt64(),
			Pooler_enabled:           data.PoolerEnabled.ValueBool(),
			Pooler_mode:              data.PoolerMode.ValueString(),
			Disabled:                 data.Disabled.ValueBool(),
			Passwordless_access:      data.PasswordlessAccess.ValueBool(),
		},
	}
	response, _ := r.client.R().
		SetBody(content).
		Post(fmt.Sprintf("/projects/%s/endpoints", data.ProjectID.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create endpoint resource with a status code: %s", response.Status()), string(response.Body()))
		return
	}
	endpoint := struct {
		Endpoint endpointResourceJSON `json:"endpoint"`
	}{}
	err := json.Unmarshal(response.Body(), &endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	data = *endpoint.Endpoint.ToEndpointResourceModel()
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r endpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data endpointResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	response, _ := r.client.R().Get(fmt.Sprintf("/projects/%s/endpoints/%s", data.ProjectID.ValueString(), data.Id.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), "")
		return
	}

	endpoint := struct {
		Endpoint endpointResourceJSON `json:"endpoint"`
	}{}
	err := json.Unmarshal(response.Body(), &endpoint)
	data = *endpoint.Endpoint.ToEndpointResourceModel()
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

type updateEndpoint struct {
	Branch_id                string                `json:"branch_id"`
	Settings                 *endpointSettingsJSON `json:"settings,omitempty"`
	Autoscaling_limit_min_cu int64                 `json:"autoscaling_limit_min_cu,omitempty"`
	Autoscaling_limit_max_cu int64                 `json:"autoscaling_limit_max_cu,omitempty"`
	Pooler_enabled           bool                  `json:"pooler_enabled,omitempty"`
	Pooler_mode              string                `json:"pooler_mode,omitempty"`
	Disabled                 bool                  `json:"disabled,omitempty"`
	Passwordless_access      bool                  `json:"passwordless_access,omitempty"`
}

func (r endpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data endpointResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	content := updateEndpoint{
		Branch_id:                data.BranchID.ValueString(),
		Autoscaling_limit_min_cu: data.AutoscalingLimitMinCu.ValueInt64(),
		Autoscaling_limit_max_cu: data.AutoscalingLimitMaxCu.ValueInt64(),
		Pooler_enabled:           data.PoolerEnabled.ValueBool(),
		Pooler_mode:              data.PoolerMode.ValueString(),
		Disabled:                 data.Disabled.ValueBool(),
		Passwordless_access:      data.PasswordlessAccess.ValueBool(),
	}
	response, _ := r.client.R().
		SetBody(content).
		Patch(fmt.Sprintf("/projects/%s/endpoints", data.ProjectID.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to update endpoint resource with a status code: %s", response.Status()), "")
		return
	}

	endpoint := struct {
		Endpoint endpointResourceJSON `json:"endpoint"`
	}{}
	err := json.Unmarshal(response.Body(), &endpoint)
	data = *endpoint.Endpoint.ToEndpointResourceModel()
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

func (r endpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data endpointResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	response, _ := r.client.R().Delete(fmt.Sprintf("/projects/%s/endpoints/%s", data.ProjectID.ValueString(), data.Id.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), "")
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r endpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
