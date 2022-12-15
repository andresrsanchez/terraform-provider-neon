package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type branchResource struct {
	client *http.Client
}
type branchResourceJSON struct {
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
}

var _ resource.Resource = branchResource{}
var _ resource.ResourceWithImportState = branchResource{}

func NewBranchResource() resource.Resource {
	return branchResource{
		client: &http.Client{},
	}
}

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
	}
}

func toBranchModel(in *branchResourceJSON) types.Object {
	branch := BranchResourceModel{
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
	aux, _ := types.ObjectValueFrom(context.TODO(), typeFromAttrs(branchResourceAttr()), branch)
	return aux
}

func toBranchJSON(in *types.Object) *branchResourceJSON {
	v := BranchResourceModel{}
	in.As(context.TODO(), &v, types.ObjectAsOptions{})
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
	}
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
		Branch createBranchInner `json:"branch"`
	}{
		Branch: createBranchInner{
			Parent_id:        data.ParentID.ValueString(),
			Name:             data.Name.ValueString(),
			Parent_lsn:       data.ParentLsn.ValueString(),
			Parent_timestamp: data.ParentLsn.ValueString(),
		},
	}
	b, err := json.Marshal(content)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal project", err.Error())
		return
	}
	key, ok := os.LookupEnv("NEON_API_KEY")
	if !ok {
		resp.Diagnostics.AddError(
			"Unable to find token",
			"token cannot be an empty string",
		)
		return
	}
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches", data.ProjectID.ValueString())
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create the request", err.Error())
		return
	}
	request.Header.Add("Authorization", "Bearer "+key)
	request.Header.Set("Content-Type", "application/json")
	response, err := r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create branch", err.Error())
		return
	} else if response.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError("Failed to create branch, response status ", response.Status)
		return
	}
	defer response.Body.Close()

	branchJSON := struct {
		Branch branchResourceJSON `json:"branch"`
	}{}
	err = json.NewDecoder(response.Body).Decode(&branchJSON)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	branchObj := toBranchModel(&branchJSON.Branch)
	diags = resp.State.Set(ctx, branchObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
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
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s", projectID.ValueString(), ID.ValueString())
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create the request", err.Error())
		return
	}
	key, ok := os.LookupEnv("NEON_API_KEY")
	if !ok {
		resp.Diagnostics.AddError(
			"Unable to find token",
			"token cannot be an empty string",
		)
		return
	}
	request.Header.Add("Authorization", "Bearer "+key)
	response, err := r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete branch", err.Error())
		return
	} else if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Failed to delete branch, response status ", response.Status)
		return
	}
	resp.State.RemoveResource(ctx)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r branchResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_branch"
}

func (r branchResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
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
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s", projectID.ValueString(), ID.ValueString())
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create the request", err.Error())
		return
	}
	key, ok := os.LookupEnv("NEON_API_KEY")
	if !ok {
		resp.Diagnostics.AddError(
			"Unable to find token",
			"token cannot be an empty string",
		)
		return
	}
	request.Header.Add("Authorization", "Bearer "+key)
	request.Header.Set("Content-Type", "application/json")
	response, err := r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create project", err.Error())
		return
	} else if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Failed to create project, response status ", response.Status)
		return
	}
	defer response.Body.Close()
	inner := struct {
		Branch branchResourceJSON `json:"branch"`
	}{}
	err = json.NewDecoder(response.Body).Decode(&inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	branchObj := toBranchModel(&inner.Branch)
	diags = resp.State.Set(ctx, branchObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r branchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data BranchResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	content := struct {
		Name string `json:"name"`
	}{
		Name: data.Name.ValueString(),
	}
	b, err := json.Marshal(content)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal endpoint", err.Error())
		return
	}
	key, ok := os.LookupEnv("NEON_API_KEY")
	if !ok {
		resp.Diagnostics.AddError(
			"Unable to find token",
			"token cannot be an empty string",
		)
		return
	}
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s", data.ProjectID.ValueString(), data.ID.ValueString())
	request, err := http.NewRequest("PATCH", url, bytes.NewBuffer(b))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create the request", err.Error())
		return
	}
	request.Header.Add("Authorization", "Bearer "+key)
	request.Header.Set("Content-Type", "application/json")
	response, err := r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update branch", err.Error())
		return
	} else if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Failed to update branch, response status ", response.Status)
		return
	}
	defer response.Body.Close()
	branchJSON := struct {
		Branch branchResourceJSON `json:"branch"`
	}{}
	err = json.NewDecoder(response.Body).Decode(&branchJSON)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	branchObj := toBranchModel(&branchJSON.Branch)
	diags = resp.State.Set(ctx, branchObj)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r branchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
