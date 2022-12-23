package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &endpointDataSource{}

type endpointDataSource struct {
	client *resty.Client
}

type endpointDataModel struct {
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

type endpointDataJSON struct {
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

func (in *endpointDataJSON) ToEndpointDataModel() *endpointDataModel {
	return &endpointDataModel{
		Host:                  types.StringValue(in.Host),
		Id:                    types.StringValue(in.Id),
		ProjectID:             types.StringValue(in.ProjectID),
		BranchID:              types.StringValue(in.BranchID),
		AutoscalingLimitMinCu: types.Int64Value(in.AutoscalingLimitMinCu),
		AutoscalingLimitMaxCu: types.Int64Value(in.AutoscalingLimitMaxCu),
		RegionID:              types.StringValue(in.RegionID),
		Type:                  types.StringValue(in.Type),
		CurrentState:          types.StringValue(in.CurrentState),
		PendingState:          types.StringValue(in.PendingState),
		PoolerEnabled:         types.BoolValue(in.PoolerEnabled),
		PoolerMode:            types.StringValue(in.PoolerMode),
		Disabled:              types.BoolValue(in.Disabled),
		PasswordlessAccess:    types.BoolValue(in.PasswordlessAccess),
		LastActive:            types.StringValue(in.LastActive),
		CreatedAt:             types.StringValue(in.CreatedAt),
		UpdatedAt:             types.StringValue(in.UpdatedAt),
	}
}

// Metadata implements datasource.DataSource
func (*endpointDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoint"
}

// Read implements datasource.DataSource
func (d *endpointDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data endpointDataModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, _ := d.client.R().Get(fmt.Sprintf("/projects/%s/endpoints/%s", data.ProjectID.ValueString(), data.Id.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), "")
		return
	}
	endpoint := struct {
		Endpoint endpointDataJSON `json:"endpoint"`
	}{}
	err := json.Unmarshal(response.Body(), &endpoint)
	data = *endpoint.Endpoint.ToEndpointDataModel()
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Schema implements datasource.DataSource
func (*endpointDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "neon host",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "endpoint id",
				Required:            true,
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "project id",
				Required:            true,
			},
			"branch_id": schema.StringAttribute{
				MarkdownDescription: "postgres branch",
				Computed:            true,
			},
			"autoscaling_limit_min_cu": schema.Int64Attribute{
				MarkdownDescription: "autoscaling limit min",
				Computed:            true,
			},
			"autoscaling_limit_max_cu": schema.Int64Attribute{
				MarkdownDescription: "autoscaling limit max",
				Computed:            true,
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "region id",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "type",
				Computed:            true,
			},
			"current_state": schema.StringAttribute{
				MarkdownDescription: "current state",
				Computed:            true,
			},
			"pending_state": schema.StringAttribute{
				MarkdownDescription: "pending state",
				Computed:            true,
			},
			"pooler_enabled": schema.BoolAttribute{
				MarkdownDescription: "pooler enabled",
				Computed:            true,
			},
			"pooler_mode": schema.StringAttribute{
				MarkdownDescription: "pooler mode",
				Computed:            true,
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "disabled",
				Computed:            true,
			},
			"passwordless_access": schema.BoolAttribute{
				MarkdownDescription: "passwordless access",
				Computed:            true,
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
		},
	}
}
