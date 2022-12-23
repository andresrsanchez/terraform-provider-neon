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
var _ datasource.DataSource = &branchDataSource{}

type branchDataSource struct {
	client *resty.Client
}

type branchDataModel struct {
	ID               types.String `tfsdk:"id"`
	ProjectID        types.String `tfsdk:"project_id"`
	ParentID         types.String `tfsdk:"parent_id"`
	ParentLsn        types.String `tfsdk:"parent_lsn"`
	Name             types.String `tfsdk:"name"`
	CurrentState     types.String `tfsdk:"current_state"`
	CreatedAt        types.String `tfsdk:"created_at"`
	UpdatedAt        types.String `tfsdk:"updated_at"`
	ParentTimestamp  types.String `tfsdk:"parent_timestamp"`
	PendingState     types.String `tfsdk:"pending_state"`
	LogicalSize      types.Int64  `tfsdk:"logical_size"`
	LogicalSizeLimit types.Int64  `tfsdk:"logical_size_limit"`
	PhysicalSize     types.Int64  `tfsdk:"physical_size"`
}

type branchDataJSON struct {
	ID               string `json:"id"`
	ProjectID        string `json:"project_id"`
	ParentID         string `json:"parent_id"`
	ParentLsn        string `json:"parent_lsn"`
	Name             string `json:"name"`
	CurrentState     string `json:"current_state"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
	ParentTimestamp  string `json:"parent_timestamp"`
	PendingState     string `json:"pending_state"`
	LogicalSize      int64  `json:"logical_size"`
	LogicalSizeLimit int64  `json:"logical_size_limit"`
	PhysicalSize     int64  `json:"physical_size"`
}

func (in *branchDataJSON) ToBranchDataModel() *branchDataModel {
	return &branchDataModel{
		ID:               types.StringValue(in.ID),
		ProjectID:        types.StringValue(in.ProjectID),
		ParentID:         types.StringValue(in.ParentID),
		ParentLsn:        types.StringValue(in.ParentLsn),
		Name:             types.StringValue(in.Name),
		CurrentState:     types.StringValue(in.CurrentState),
		CreatedAt:        types.StringValue(in.CreatedAt),
		UpdatedAt:        types.StringValue(in.UpdatedAt),
		ParentTimestamp:  types.StringValue(in.ParentTimestamp),
		PendingState:     types.StringValue(in.PendingState),
		LogicalSize:      types.Int64Value(in.LogicalSize),
		LogicalSizeLimit: types.Int64Value(in.LogicalSizeLimit),
		PhysicalSize:     types.Int64Value(in.PhysicalSize),
	}
}

// Metadata implements datasource.DataSource
func (*branchDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_branch"
}

// Read implements datasource.DataSource
func (d *branchDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data branchDataModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, _ := d.client.R().Get(fmt.Sprintf("/projects/%s/branches/%s", data.ProjectID.ValueString(), data.ID.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), "")
		return
	}
	branch := struct {
		Branch branchDataJSON `json:"branch"`
	}{}
	err := json.Unmarshal(response.Body(), &branch)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}

	plan := branch.Branch.ToBranchDataModel()
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Schema implements datasource.DataSource
func (*branchDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"project_id": schema.StringAttribute{
				Required: true,
			},
			"parent_id": schema.StringAttribute{
				Computed: true,
			},
			"parent_lsn": schema.StringAttribute{
				Computed: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"current_state": schema.StringAttribute{
				Computed: true,
			},
			"created_at": schema.StringAttribute{
				Computed: true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
			},
			"parent_timestamp": schema.StringAttribute{
				Computed: true,
			},
			"pending_state": schema.StringAttribute{
				Computed: true,
			},
			"logical_size": schema.Int64Attribute{
				Computed: true,
			},
			"logical_size_limit": schema.Int64Attribute{
				Computed: true,
			},
			"physical_size": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}
