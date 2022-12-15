package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type roleResource struct {
	client *http.Client
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

func NewRoleResource() resource.Resource {
	return roleResource{
		client: &http.Client{},
	}
}

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

func toRoleJSON(in *types.Object) *roleResourceJSON {
	v := roleResourceModel{}
	in.As(context.TODO(), &in, types.ObjectAsOptions{})
	return &roleResourceJSON{
		BranchID:  v.BranchID.ValueString(),
		Name:      v.Name.ValueString(),
		Password:  v.Password.ValueString(),
		Protected: v.Protected.ValueBool(),
		CreatedAt: v.UpdatedAt.ValueString(),
		UpdatedAt: v.UpdatedAt.ValueString(),
	}
}

func toRoleModel(v *roleResourceJSON) types.Object {
	db := roleResourceModel{
		BranchID:  types.StringValue(v.BranchID),
		Name:      types.StringValue(v.Name),
		Password:  types.StringValue(v.Password),
		Protected: types.BoolValue(v.Protected),
		CreatedAt: types.StringValue(v.CreatedAt),
		UpdatedAt: types.StringValue(v.UpdatedAt),
	}
	aux, _ := types.ObjectValueFrom(context.TODO(), typeFromAttrs(databaseResourceAttr()), db)
	return aux
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
	b, err := json.Marshal(content)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal role", err.Error())
		return
	}
	fmt.Println(string(b))
	key, ok := os.LookupEnv("NEON_API_KEY")
	if !ok {
		resp.Diagnostics.AddError(
			"Unable to find token",
			"token cannot be an empty string",
		)
		return
	}
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s/roles", data.ProjectID.ValueString(), data.BranchID.ValueString())
	fmt.Println(url)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create the request", err.Error())
		return
	}
	request.Header.Add("Authorization", "Bearer "+key)
	request.Header.Set("Content-Type", "application/json")
	response, err := r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create role", err.Error())
		return
	} else if response.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError("Failed to create role, response status ", response.Status)
		return
	}
	defer response.Body.Close()
	inner := struct {
		Role roleResourceJSON `json:"role"`
	}{
		Role: roleResourceJSON{},
	}
	err = json.NewDecoder(response.Body).Decode(&inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	diags = resp.State.Set(ctx, toRoleModel(&inner.Role))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
}

func (r roleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data roleResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s/roles/%s", data.ProjectID.ValueString(), data.BranchID.ValueString(), data.Name.ValueString())
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
		resp.Diagnostics.AddError("Failed to delete role", err.Error())
		return
	} else if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Failed to delete role, response status ", response.Status)
		return
	}
	resp.State.RemoveResource(ctx)
	if resp.Diagnostics.HasError() {
		return
	}
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
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s/roles/%s", data.ProjectID.ValueString(), data.BranchID.ValueString(), data.Name.ValueString())
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
		Role roleResourceJSON `json:"role"`
	}{
		Role: roleResourceJSON{},
	}
	err = json.NewDecoder(response.Body).Decode(&inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	diags = resp.State.Set(ctx, toRoleModel(&inner.Role))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
}

func (r roleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	time.Sleep(20 * time.Second)
	var data roleResourceModel
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
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
	//review
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s/roles/%s/reset_password", data.ProjectID.ValueString(), data.BranchID.ValueString(), data.Name.ValueString())
	fmt.Println(url)
	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create the request", err.Error())
		return
	}
	request.Header.Add("Authorization", "Bearer "+key)
	request.Header.Set("Content-Type", "application/json")
	response, err := r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create role", err.Error())
		return
	} else if response.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError("Failed to create role, response status ", response.Status)
		return
	}
	defer response.Body.Close()
	inner := struct {
		Role roleResourceJSON `json:"role"`
	}{
		Role: roleResourceJSON{},
	}
	err = json.NewDecoder(response.Body).Decode(&inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	diags = resp.State.Set(ctx, toRoleModel(&inner.Role))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
}

func (r roleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
