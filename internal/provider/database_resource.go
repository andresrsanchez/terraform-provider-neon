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

type databaseResource struct {
	client *resty.Client
}
type databaseResourceJSON struct {
	ID        int64  `json:"id"`
	BranchID  string `json:"branch_id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	OwnerName string `json:"owner_name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
type databaseResourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	BranchID  types.String `tfsdk:"branch_id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	OwnerName types.String `tfsdk:"owner_name"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

var _ resource.Resource = databaseResource{}
var _ resource.ResourceWithImportState = databaseResource{}

func databaseResourceAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed: true,
		},
		"branch_id": schema.StringAttribute{
			Required: true,
		},
		"project_id": schema.StringAttribute{
			Required: true,
		},
		"name": schema.StringAttribute{
			Required: true,
		},
		"owner_name": schema.StringAttribute{
			Optional: true,
		},
		"created_at": schema.StringAttribute{
			Computed: true,
		},
		"updated_at": schema.StringAttribute{
			Computed: true,
		},
	}
}

func toDatabaseJSON(v *types.Object) (*databaseResourceJSON, diag.Diagnostics) {
	in := databaseResourceModel{}
	diags := v.As(context.TODO(), &in, types.ObjectAsOptions{})
	return &databaseResourceJSON{
		ID:        in.ID.ValueInt64(),
		BranchID:  in.BranchID.ValueString(),
		Name:      in.Name.ValueString(),
		OwnerName: in.OwnerName.ValueString(),
		CreatedAt: in.CreatedAt.ValueString(),
		UpdatedAt: in.UpdatedAt.ValueString(),
	}, diags
}

func toDatabaseModel(in *databaseResourceJSON) (types.Object, diag.Diagnostics) {
	db := databaseResourceModel{
		ID:        types.Int64Value(in.ID),
		BranchID:  types.StringValue(in.BranchID),
		Name:      types.StringValue(in.Name),
		OwnerName: types.StringValue(in.OwnerName),
		CreatedAt: types.StringValue(in.CreatedAt),
		UpdatedAt: types.StringValue(in.UpdatedAt),
	}
	return types.ObjectValueFrom(context.TODO(), typeFromAttrs(databaseResourceAttr()), db)
}

func (r databaseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: databaseResourceAttr(),
	}
}

func (r databaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	time.Sleep(20 * time.Second)
	var data databaseResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	content := struct {
		Database struct {
			Name      string `json:"name"`
			OwnerName string `json:"owner_name,omitempty"`
		} `json:"database"`
	}{
		Database: struct {
			Name      string `json:"name"`
			OwnerName string `json:"owner_name,omitempty"`
		}{
			Name:      data.Name.ValueString(),
			OwnerName: data.OwnerName.ValueString(),
		},
	}
	url := fmt.Sprintf("/projects/%s/branches/%s/databases", data.ProjectID.ValueString(), data.BranchID.ValueString())
	response, err := r.client.R().
		SetBody(content).
		Post(url)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create endpoint resource with a status code: %s", response.Status()), err.Error())
		return
	}

	inner := struct {
		Database databaseResourceJSON `json:"database"`
	}{
		Database: databaseResourceJSON{},
	}
	err = json.Unmarshal(response.Body(), &inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}

	databaseObj, diags := toDatabaseModel(&inner.Database)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
	diags = resp.State.Set(ctx, databaseObj)
	resp.Diagnostics.Append(diags...)
}

func (r databaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data databaseResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.R().Delete(fmt.Sprintf("/projects/%s/branches/%s/databases/%s", data.ProjectID.ValueString(), data.BranchID.ValueString(), data.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r databaseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r databaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data databaseResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.R().Get(fmt.Sprintf("/projects/%s/branches/%s/databases/%s", data.ProjectID.ValueString(), data.BranchID.ValueString(), data.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), err.Error())
		return
	}
	inner := struct {
		Database databaseResourceJSON `json:"database"`
	}{
		Database: databaseResourceJSON{},
	}
	err = json.Unmarshal(response.Body(), &inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	databaseObj, diags := toDatabaseModel(&inner.Database)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
	diags = resp.State.Set(ctx, databaseObj)
	resp.Diagnostics.Append(diags...)
}

func (r databaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	time.Sleep(20 * time.Second)
	var data databaseResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	content := struct {
		Database struct {
			Name      string `json:"name"`
			OwnerName string `json:"owner_name,omitempty"`
		} `json:"database"`
	}{
		Database: struct {
			Name      string `json:"name"`
			OwnerName string `json:"owner_name,omitempty"`
		}{
			Name:      data.Name.ValueString(),
			OwnerName: data.OwnerName.ValueString(),
		},
	}
	response, err := r.client.R().
		SetBody(content).
		Patch(fmt.Sprintf("/projects/%s/branches/%s/databases/%s", data.ProjectID.ValueString(), data.BranchID.ValueString(), data.Name.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create endpoint resource with a status code: %s", response.Status()), err.Error())
		return
	}

	inner := struct {
		Database databaseResourceJSON `json:"database"`
	}{
		Database: databaseResourceJSON{},
	}
	err = json.Unmarshal(response.Body(), &inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	databaseObj, diags := toDatabaseModel(&inner.Database)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
	diags = resp.State.Set(ctx, databaseObj)
	resp.Diagnostics.Append(diags...)
}

func (r databaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
