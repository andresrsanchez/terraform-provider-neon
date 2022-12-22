package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type branchResource struct {
	client *resty.Client
}
type branchResourceEndpointJSON struct {
	Type                  string `json:"type"`
	AutoscalingLimitMinCu int64  `json:"autoscaling_limit_min_cu"`
	AutoscalingLimitMaxCu int64  `json:"autoscaling_limit_max_cu"`
}
type branchResourceJSON struct {
	ID               string                 `json:"id"`
	ProjectID        string                 `json:"project_id"`
	ParentID         string                 `json:"parent_id"`
	ParentLsn        string                 `json:"parent_lsn"`
	Name             string                 `json:"name"`
	CurrentState     string                 `json:"current_state"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
	ParentTimestamp  string                 `json:"parent_timestamp"`
	PendingState     string                 `json:"pending_state"`
	LogicalSize      int64                  `json:"logical_size"`
	LogicalSizeLimit int64                  `json:"logical_size_limit"`
	PhysicalSize     int64                  `json:"physical_size"`
	Endpoints        []endpointResourceJSON `json:"endpoints"`
}

func (in *branchResourceJSON) ToBranchResourceModel(ctx context.Context) (*BranchResourceModel, diag.Diagnostics) {
	branch := &BranchResourceModel{
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
		Endpoints:        types.ListNull(types.ObjectType{AttrTypes: typeFromAttrs(branchResourceEndpointAttr())}),
	}
	if len(in.Endpoints) > 0 {
		e := []endpointResourceModel{}
		for _, v := range in.Endpoints {
			e = append(e, *v.ToEndpointResourceModel())
		}
		aux, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: typeFromAttrs(branchResourceEndpointAttr())}, e)
		if diags.HasError() {
			return nil, diags
		}
		branch.Endpoints = aux
	}
	return branch, nil
}

func (in *BranchResourceModel) ToBranchResourceObject(ctx context.Context) (types.Object, diag.Diagnostics) {
	return types.ObjectValueFrom(ctx, typeFromAttrs(branchResourceAttr()), in)
}

type BranchResourceModel struct {
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
	Endpoints        types.List   `tfsdk:"endpoints"`
}

var _ resource.Resource = branchResource{}
var _ resource.ResourceWithImportState = branchResource{}

func branchResourceAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"project_id": schema.StringAttribute{
			//Computed: true,
			Required: true,
		},
		"parent_id": schema.StringAttribute{
			Computed: true,
			//Required: true,
		},
		"parent_lsn": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Optional: true,
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
		"endpoints": schema.ListNestedAttribute{
			NestedObject: schema.NestedAttributeObject{
				Attributes: branchResourceEndpointAttr(),
			},
			Optional: true,
			//Computed: true,
			//Required: true,
			// PlanModifiers: []planmodifier.List{
			// 	listplanmodifier.UseStateForUnknown(),
			// },
		},
	}
}

func branchResourceEndpointAttr() map[string]schema.Attribute {
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
			Computed:            true,
		},
		"branch_id": schema.StringAttribute{
			MarkdownDescription: "postgres branch",
			Computed:            true,
			//PlanModifiers: planmodifier.str,
		},
		"autoscaling_limit_min_cu": schema.Int64Attribute{
			MarkdownDescription: "autoscaling limit min",
			Required:            true,
		},
		"autoscaling_limit_max_cu": schema.Int64Attribute{
			MarkdownDescription: "autoscaling limit max",
			Required:            true,
		},
		"region_id": schema.StringAttribute{
			MarkdownDescription: "region id",
			Computed:            true,
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
		},
		"pooler_mode": schema.StringAttribute{
			MarkdownDescription: "pooler mode",
			Computed:            true,
			Validators:          []validator.String{stringvalidator.OneOf("transaction")},
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
	}
}

func toBranchJSON(in *types.Object, ctx context.Context) (*branchResourceJSON, diag.Diagnostics) {
	v := BranchResourceModel{
		Endpoints: types.ListNull(types.ObjectType{AttrTypes: typeFromAttrs(branchResourceEndpointAttr())}),
	}
	diags := in.As(ctx, &v, basetypes.ObjectAsOptions{})
	endpoints := []endpointResourceJSON{}
	if !v.Endpoints.IsNull() {
		for _, vv := range v.Endpoints.Elements() {
			endpointModel := endpointResourceModel{}
			diags := vv.(types.Object).As(ctx, &endpointModel, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return nil, diags
			}
			endpoint := *endpointModel.ToEndpointResourceJSON()
			endpoints = append(endpoints, endpoint)
		}
	}
	return &branchResourceJSON{
		ID:               v.ID.ValueString(),
		ProjectID:        v.ProjectID.ValueString(),
		ParentID:         v.ParentID.ValueString(),
		ParentLsn:        v.ParentLsn.ValueString(),
		Name:             v.Name.ValueString(),
		CurrentState:     v.CurrentState.ValueString(),
		CreatedAt:        v.CreatedAt.ValueString(),
		UpdatedAt:        v.UpdatedAt.ValueString(),
		ParentTimestamp:  v.ParentTimestamp.ValueString(),
		PendingState:     v.PendingState.ValueString(),
		LogicalSize:      v.LogicalSize.ValueInt64(),
		LogicalSizeLimit: v.LogicalSizeLimit.ValueInt64(),
		PhysicalSize:     v.PhysicalSize.ValueInt64(),
		Endpoints:        endpoints,
	}, diags
}

func (r branchResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: branchResourceAttr(),
	}
}

type createBranchInner struct {
	Parent_id        string `json:"parent_id,omitempty"`
	Name             string `json:"name,omitempty"`
	Parent_lsn       string `json:"parent_lsn,omitempty"`
	Parent_timestamp string `json:"parent_timestamp,omitempty"`
}

func (r branchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data BranchResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	content := struct {
		Branch    createBranchInner            `json:"branch"`
		Endpoints []branchResourceEndpointJSON `json:"endpoints"`
	}{
		Branch: createBranchInner{
			Parent_id:        data.ParentID.ValueString(),
			Name:             data.Name.ValueString(),
			Parent_lsn:       data.ParentLsn.ValueString(),
			Parent_timestamp: data.ParentTimestamp.ValueString(),
		},
		Endpoints: []branchResourceEndpointJSON{},
	}

	for _, vv := range data.Endpoints.Elements() {
		endpoint := endpointResourceModel{}
		diags := vv.(types.Object).As(ctx, &endpoint, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		content.Endpoints = append(content.Endpoints, branchResourceEndpointJSON{
			Type:                  endpoint.Type.ValueString(),
			AutoscalingLimitMinCu: endpoint.AutoscalingLimitMinCu.ValueInt64(),
			AutoscalingLimitMaxCu: endpoint.AutoscalingLimitMaxCu.ValueInt64(),
		})
	}
	response, _ := r.client.R().
		EnableTrace().
		SetBody(content).
		Post(fmt.Sprintf("/projects/%s/branches", data.ProjectID.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create endpoint resource with a status code: %s", response.Status()), "")
		return
	}

	branchJSON := struct {
		Branch    branchResourceJSON     `json:"branch"`
		Endpoints []endpointResourceJSON `json:"endpoints"`
	}{}
	err := json.Unmarshal(response.Body(), &branchJSON)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	branchJSON.Branch.Endpoints = branchJSON.Endpoints
	branchObj, diags := branchJSON.Branch.ToBranchResourceModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, branchObj)
	resp.Diagnostics.Append(diags...)
}

func (r branchResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var projectID, ID types.String
	diags := req.State.GetAttribute(ctx, path.Root("project_id"), &projectID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = req.State.GetAttribute(ctx, path.Root("id"), &ID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	response, _ := r.client.R().Delete(fmt.Sprintf("/projects/%s/branches/%s", projectID.ValueString(), ID.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), "")
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r branchResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_branch"
}

func (r branchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BranchResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	response, _ := r.client.R().Get(fmt.Sprintf("/projects/%s/branches/%s", data.ProjectID.ValueString(), data.ID.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), "")
		return
	}
	bodyBranch := struct {
		Branch branchResourceJSON `json:"branch"`
	}{}
	err := json.Unmarshal(response.Body(), &bodyBranch)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	branchObj, diags := bodyBranch.Branch.ToBranchResourceModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, _ = r.client.R().Get(fmt.Sprintf("/projects/%s/branches/%s/endpoints", data.ProjectID.ValueString(), data.ID.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), "")
		return
	}
	bodyEndpoints := struct {
		Endpoints []endpointResourceJSON
	}{}
	err = json.Unmarshal(response.Body(), &bodyEndpoints)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	if len(bodyEndpoints.Endpoints) > 0 {
		e := []endpointResourceModel{}
		for _, v := range bodyEndpoints.Endpoints {
			e = append(e, *v.ToEndpointResourceModel())
		}
		aux, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: typeFromAttrs(branchResourceEndpointAttr())}, e)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		branchObj.Endpoints = aux
	}

	diags = resp.State.Set(ctx, branchObj)
	resp.Diagnostics.Append(diags...)
}

func (r branchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BranchResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state BranchResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	content := struct {
		Branch struct {
			Name string `json:"name"`
		} `json:"branch"`
	}{
		Branch: struct {
			Name string `json:"name"`
		}{
			Name: data.Name.ValueString(),
		},
	}
	response, _ := r.client.R().
		SetBody(content).
		Patch(fmt.Sprintf("/projects/%s/branches/%s", state.ProjectID.ValueString(), state.ID.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create endpoint resource with a status code: %s", response.Status()), "")
		return
	}
	branchJSON := struct {
		Branch branchResourceJSON `json:"branch"`
	}{}
	err := json.Unmarshal(response.Body(), &branchJSON)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	branchModel, diags := branchJSON.Branch.ToBranchResourceModel(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, _ = r.client.R().Get(fmt.Sprintf("/projects/%s/branches/%s/endpoints", state.ProjectID.ValueString(), state.ID.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), "")
		return
	}
	bodyEndpoints := struct {
		Endpoints []endpointResourceJSON
	}{}
	err = json.Unmarshal(response.Body(), &bodyEndpoints)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	if len(bodyEndpoints.Endpoints) > 0 {
		e := []endpointResourceModel{}
		for _, v := range bodyEndpoints.Endpoints {
			e = append(e, *v.ToEndpointResourceModel())
		}
		aux, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: typeFromAttrs(branchResourceEndpointAttr())}, e)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		branchModel.Endpoints = aux
	}
	diags = resp.State.Set(ctx, branchModel)
	resp.Diagnostics.Append(diags...)
}

func (r branchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ids := strings.Split(req.ID, "/")
	if len(ids) != 2 || ids[0] == "" || ids[1] == "" {
		resp.Diagnostics.AddError(
			"Cannot import branch",
			fmt.Sprintf("Invalid id '%s' specified. should be in format \"id/project_id\"", req.ID),
		)
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), ids[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), ids[1])...)
}
