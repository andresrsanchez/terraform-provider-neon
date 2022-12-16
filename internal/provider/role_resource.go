package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type roleResource struct {
	client *resty.Client
}
type roleResourceJSON struct {
	ProjectID string `json:"project_id"`
	BranchID  string `json:"branch_id"`
	Name      string `json:"name"`
	Password  string `json:"password"`
	Protected bool   `json:"protected"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
type roleResourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	BranchID  types.String `tfsdk:"branch_id"`
	Name      types.String `tfsdk:"name"`
	Password  types.String `tfsdk:"password"`
	Protected types.Bool   `tfsdk:"protected"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

var _ resource.Resource = roleResource{}
var _ resource.ResourceWithImportState = roleResource{}

func roleResourceAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"branch_id": schema.StringAttribute{
			Required: true,
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"project_id": schema.StringAttribute{
			Required: true,
		},
		"password": schema.StringAttribute{
			Computed: true,
		},
		"protected": schema.BoolAttribute{
			Computed: true,
		},
		"created_at": schema.StringAttribute{
			Computed: true,
		},
		"updated_at": schema.StringAttribute{
			Computed: true,
		},
	}
}

func toRoleJSON(in *types.Object) (*roleResourceJSON, diag.Diagnostics) {
	v := roleResourceModel{}
	diags := in.As(context.TODO(), &in, types.ObjectAsOptions{})
	return &roleResourceJSON{
		BranchID:  v.BranchID.ValueString(),
		Name:      v.Name.ValueString(),
		Password:  v.Password.ValueString(),
		Protected: v.Protected.ValueBool(),
		CreatedAt: v.UpdatedAt.ValueString(),
		UpdatedAt: v.UpdatedAt.ValueString(),
	}, diags
}

func toRoleModel(v *roleResourceJSON) (types.Object, diag.Diagnostics) {
	db := roleResourceModel{
		BranchID:  types.StringValue(v.BranchID),
		Name:      types.StringValue(v.Name),
		Password:  types.StringValue(v.Password),
		Protected: types.BoolValue(v.Protected),
		CreatedAt: types.StringValue(v.CreatedAt),
		UpdatedAt: types.StringValue(v.UpdatedAt),
	}
	return types.ObjectValueFrom(context.TODO(), typeFromAttrs(databaseResourceAttr()), db)
}

func (r roleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: databaseResourceAttr(),
	}
}

func (r roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	time.Sleep(20 * time.Second)
	var data roleResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	content := struct {
		Role struct {
			Name string `json:"name"`
		} `json:"role"`
	}{
		Role: struct {
			Name string `json:"name"`
		}{
			Name: data.Name.ValueString(),
		},
	}
	response, err := r.client.R().
		SetBody(content).
		Post(fmt.Sprintf("/projects/%s/branches/%s/roles", data.ProjectID.ValueString(), data.BranchID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create role resource with a status code: %s", response.Status()), err.Error())
		return
	}

	inner := struct {
		Role roleResourceJSON `json:"role"`
	}{
		Role: roleResourceJSON{},
	}
	err = json.Unmarshal(response.Body(), &inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	roleObj, diags := toRoleModel(&inner.Role)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
	diags = resp.State.Set(ctx, roleObj)
	resp.Diagnostics.Append(diags...)
}

func (r roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data roleResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.R().Delete(fmt.Sprintf("/projects/%s/branches/%s/roles/%s", data.ProjectID.ValueString(), data.BranchID.ValueString(), data.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r roleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r roleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data roleResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.R().Get(fmt.Sprintf("/projects/%s/branches/%s/roles/%s", data.ProjectID.ValueString(), data.BranchID.ValueString(), data.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), err.Error())
		return
	}
	inner := struct {
		Role roleResourceJSON `json:"role"`
	}{
		Role: roleResourceJSON{},
	}
	err = json.Unmarshal(response.Body(), &inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	roleObj, diags := toRoleModel(&inner.Role)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
	diags = resp.State.Set(ctx, roleObj)
	resp.Diagnostics.Append(diags...)
}

func (r roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	time.Sleep(20 * time.Second)
	var data roleResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, err := r.client.R().
		Post(fmt.Sprintf("/projects/%s/branches/%s/roles/%s/reset_password", data.ProjectID.ValueString(), data.BranchID.ValueString(), data.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create role resource with a status code: %s", response.Status()), err.Error())
		return
	}

	inner := struct {
		Role roleResourceJSON `json:"role"`
	}{
		Role: roleResourceJSON{},
	}
	err = json.Unmarshal(response.Body(), &inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	roleObj, diags := toRoleModel(&inner.Role)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
	diags = resp.State.Set(ctx, roleObj)
	resp.Diagnostics.Append(diags...)
}

func (r roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
