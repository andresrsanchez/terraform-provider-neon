package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
type BranchEndpointResourceModel struct {
	Type                  types.String `tfsdk:"type"`
	AutoscalingLimitMinCu types.Int64  `tfsdk:"autoscaling_limit_min_cu"`
	AutoscalingLimitMaxCu types.Int64  `tfsdk:"autoscaling_limit_max_cu"`
}

type branchEndpointResourceJSON struct {
	Branch    branchResourceJSON `json:"branch"`
	Endpoints []struct {
		Type                  string `json:"type"`
		AutoscalingLimitMinCu int64  `json:"autoscaling_limit_min_cu"`
		AutoscalingLimitMaxCu int64  `json:"autoscaling_limit_max_cu"`
	} `json:"endpoints"`
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

func branchEndpointAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"type": schema.StringAttribute{
			MarkdownDescription: "type",
			Validators:          []validator.String{stringvalidator.OneOf("read_write", "read_only")},
			Computed:            true,
			Optional:            true,
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
	}
}

func branchEndpointResourceAttr() map[string]schema.Attribute {
	branch := branchResourceAttr()
	branch["endpoints"] = schema.ListNestedAttribute{
		Computed: true,
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: branchEndpointAttr(),
		},
	}
	return branch
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
		Attributes: branchEndpointResourceAttr(),
	}
}

type createBranchInner struct {
	Parent_id        string `json:"parent_id,omitempty"`
	Name             string `json:"name,omitempty"`
	Parent_lsn       string `json:"parent_lsn,omitempty"`
	Parent_timestamp string `json:"parent_timestamp,omitempty"`
}
type createBranchInnerEndpoint struct {
	Type                     string `json:"type,omitempty"`
	Autoscaling_limit_min_cu int64  `json:"autoscaling_limit_min_cu,omitempty"`
	Autoscaling_limit_max_cu int64  `json:"autoscaling_limit_max_cu,omitempty"`
}

func newBranchResourceModel() BranchEndpointResourceModel {
	return BranchEndpointResourceModel{
		//Object:    types.ObjectNull(typeFromAttrs(branchResourceAttr())),
		//Endpoints: types.ListNull(types.ObjectType{AttrTypes: typeFromAttrs(endpointResourceAttr())}),
	}
}

func (r branchResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	time.Sleep(20 * time.Second) //fix
	branchEndpointResourceRaw := types.ObjectNull(typeFromAttrs(branchEndpointResourceAttr()))
	diags := req.Plan.Get(ctx, &branchEndpointResourceRaw)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	endpointModelRaw := map[string]attr.Value{
		"endpoints": types.ListNull(types.ObjectType{AttrTypes: typeFromAttrs(endpointResourceAttr())}),
	}
	branchModelRaw := map[string]attr.Value{}
	for k, v := range branchEndpointResourceRaw.Attributes() {
		if k == "endpoints" {
			endpointModelRaw[k] = v
		} else {
			branchModelRaw[k] = v
		}
	}
	branchObj, diags := types.ObjectValue(typeFromAttrs(branchResourceAttr()), branchModelRaw)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var branchModel BranchResourceModel
	diags = branchObj.As(ctx, &branchModel, types.ObjectAsOptions{})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var endpointsModel1 []createBranchInnerEndpoint
	if key, ok := endpointModelRaw["endpoints"]; ok && !key.IsUnknown() { //isunknown
		var endpointsModel []BranchEndpointResourceModel
		keyList := key.(types.List)
		diags = keyList.ElementsAs(ctx, &endpointsModel, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, v := range endpointsModel {
			endpointsModel1 = append(endpointsModel1, createBranchInnerEndpoint{
				Type:                     v.Type.ValueString(),
				Autoscaling_limit_min_cu: v.AutoscalingLimitMinCu.ValueInt64(),
				Autoscaling_limit_max_cu: v.AutoscalingLimitMaxCu.ValueInt64(),
			})
		}
	}
	content := struct {
		Branch    createBranchInner           `json:"branch"`
		Endpoints []createBranchInnerEndpoint `json:"endpoints,omitempty"`
	}{
		Branch: createBranchInner{
			Parent_id:        branchModel.ParentID.ValueString(),
			Name:             branchModel.Name.ValueString(),
			Parent_lsn:       branchModel.ParentLsn.ValueString(),
			Parent_timestamp: branchModel.ParentLsn.ValueString(),
		},
		Endpoints: endpointsModel1,
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
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches", branchModel.ProjectID.ValueString())
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create the request", err.Error())
		return
	}
	request.Header.Add("Authorization", "Bearer "+key)
	request.Header.Set("Content-Type", "application/json")
	response, err := r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete branch", err.Error())
		return
	} else if response.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError("Failed to delete branch, response status ", response.Status)
		return
	}
	defer response.Body.Close()
	inner := branchEndpointResourceJSON{}
	err = json.NewDecoder(response.Body).Decode(&inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}

	branchObj = toBranchModel(&inner.Branch)
	branchEndpointRawTypes := branchObj.AttributeTypes(ctx)
	branchEndpointRawValues := branchObj.Attributes()
	branchEndpointRawTypes["endpoints"] = types.ListType{
		ElemType: types.ObjectType{AttrTypes: typeFromAttrs(endpointResourceAttr())},
	}
	branchEndpointRawValues["endpoints"] = types.ListNull(types.ObjectType{AttrTypes: typeFromAttrs(endpointResourceAttr())})

	if len(inner.Endpoints) > 0 {
		values := []BranchEndpointResourceModel{}
		for _, v := range inner.Endpoints {
			values = append(values, BranchEndpointResourceModel{
				Type:                  types.StringValue(v.Type),
				AutoscalingLimitMinCu: types.Int64Value(v.AutoscalingLimitMinCu),
				AutoscalingLimitMaxCu: types.Int64Value(v.AutoscalingLimitMaxCu),
			})
		}
		rawEndpoints, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: typeFromAttrs(endpointResourceAttr())}, values)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		branchEndpointRawValues["endpoints"] = rawEndpoints
	}

	plan, diags := types.ObjectValue(branchEndpointRawTypes, branchEndpointRawValues)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	diags = resp.State.Set(ctx, plan)
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
		fmt.Println("where is the project_id?")
		return
	}
	diags = req.State.GetAttribute(ctx, path.Root("id"), &ID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		fmt.Println("where is the id?")
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
}

func (r branchResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (r branchResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
