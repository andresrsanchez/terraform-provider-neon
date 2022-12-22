package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type roleResource struct {
	client *resty.Client
}
type roleResourceJSON struct {
	ProjectID string `json:"project_id"`
	BranchID  string `json:"branch_id"`
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
type roleResourceModel struct {
	ProjectID types.String `tfsdk:"project_id"`
	BranchID  types.String `tfsdk:"branch_id"`
	Name      types.String `tfsdk:"name"`
	Protected types.Bool   `tfsdk:"protected"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
	ID        types.String `tfsdk:"id"`
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
		"id": schema.StringAttribute{
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

func toRoleJSON(in *types.Object, ctx context.Context) (*roleResourceJSON, diag.Diagnostics) {
	v := roleResourceModel{}
	diags := in.As(ctx, &in, basetypes.ObjectAsOptions{})
	return &roleResourceJSON{
		BranchID:  v.BranchID.ValueString(),
		Name:      v.Name.ValueString(),
		Protected: v.Protected.ValueBool(),
		CreatedAt: v.UpdatedAt.ValueString(),
		UpdatedAt: v.UpdatedAt.ValueString(),
		ProjectID: v.ProjectID.ValueString(),
	}, diags
}

func toRoleModel(v *roleResourceJSON, ctx context.Context, project_id string) (types.Object, diag.Diagnostics) {
	db := roleResourceModel{
		BranchID:  types.StringValue(v.BranchID),
		Name:      types.StringValue(v.Name),
		Protected: types.BoolValue(v.Protected),
		CreatedAt: types.StringValue(v.CreatedAt),
		UpdatedAt: types.StringValue(v.UpdatedAt),
		ProjectID: types.StringValue(project_id),
		ID:        types.StringValue(v.Name),
	}
	return types.ObjectValueFrom(ctx, typeFromAttrs(roleResourceAttr()), db)
}

func (r roleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: roleResourceAttr(),
	}
}

func (r roleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
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
	response, _ := r.client.R().
		SetBody(content).
		Post(fmt.Sprintf("/projects/%s/branches/%s/roles", data.ProjectID.ValueString(), data.BranchID.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create role resource with a status code: %s", response.Status()), "")
		return
	}

	inner := struct {
		Role roleResourceJSON `json:"role"`
	}{
		Role: roleResourceJSON{},
	}
	err := json.Unmarshal(response.Body(), &inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	roleObj, diags := toRoleModel(&inner.Role, ctx, data.ProjectID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
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
	response, _ := r.client.R().Delete(fmt.Sprintf("/projects/%s/branches/%s/roles/%s", data.ProjectID.ValueString(), data.BranchID.ValueString(), data.Name.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), "")
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

	response, _ := r.client.R().Get(fmt.Sprintf("/projects/%s/branches/%s/roles/%s", data.ProjectID.ValueString(), data.BranchID.ValueString(), data.Name.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), "")
		return
	}
	inner := struct {
		Role roleResourceJSON `json:"role"`
	}{
		Role: roleResourceJSON{},
	}
	err := json.Unmarshal(response.Body(), &inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	roleObj, diags := toRoleModel(&inner.Role, ctx, data.ProjectID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
	diags = resp.State.Set(ctx, roleObj)
	resp.Diagnostics.Append(diags...)
}

func (r roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddWarning("Update role resource not implemented", "")
}

func (r roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ids := strings.Split(req.ID, "/")
	if len(ids) != 3 || ids[0] == "" || ids[1] == "" || ids[2] == "" {
		resp.Diagnostics.AddError(
			"Cannot import branch",
			fmt.Sprintf("Invalid id '%s' specified. should be in format \"project_id/branch_id/name\"", req.ID),
		)
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), ids[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("branch_id"), ids[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), ids[2])...)
}
