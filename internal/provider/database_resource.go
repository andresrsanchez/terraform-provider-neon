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

type databaseResource struct {
	client *http.Client
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

func NewDatabaseResource() resource.Resource {
	return databaseResource{
		client: &http.Client{},
	}
}

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

func toDatabaseJSON(v *types.Object) *databaseResourceJSON {
	in := databaseResourceModel{}
	v.As(context.TODO(), &in, types.ObjectAsOptions{})
	return &databaseResourceJSON{
		ID:        in.ID.ValueInt64(),
		BranchID:  in.BranchID.ValueString(),
		Name:      in.Name.ValueString(),
		OwnerName: in.OwnerName.ValueString(),
		CreatedAt: in.CreatedAt.ValueString(),
		UpdatedAt: in.UpdatedAt.ValueString(),
	}
}

func toDatabaseModel(in *databaseResourceJSON) types.Object {
	db := databaseResourceModel{
		ID:        types.Int64Value(in.ID),
		BranchID:  types.StringValue(in.BranchID),
		Name:      types.StringValue(in.Name),
		OwnerName: types.StringValue(in.OwnerName),
		CreatedAt: types.StringValue(in.CreatedAt),
		UpdatedAt: types.StringValue(in.UpdatedAt),
	}
	aux, _ := types.ObjectValueFrom(context.TODO(), typeFromAttrs(databaseResourceAttr()), db)
	return aux
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
	b, err := json.Marshal(content)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal project", err.Error())
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
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s/databases", data.ProjectID.ValueString(), data.BranchID.ValueString())
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
		resp.Diagnostics.AddError("Failed to create database", err.Error())
		return
	} else if response.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError("Failed to create database, response status ", response.Status)
		return
	}
	defer response.Body.Close()
	inner := struct {
		Database databaseResourceJSON `json:"database"`
	}{
		Database: databaseResourceJSON{},
	}
	err = json.NewDecoder(response.Body).Decode(&inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	diags = resp.State.Set(ctx, toDatabaseModel(&inner.Database))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
}

func (r databaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
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
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s/databases/%s", data.ProjectID.ValueString(), data.BranchID.ValueString(), data.Name.ValueString())
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
		Database databaseResourceJSON `json:"database"`
	}{
		Database: databaseResourceJSON{},
	}
	err = json.NewDecoder(response.Body).Decode(&inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	diags = resp.State.Set(ctx, toDatabaseModel(&inner.Database))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
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
	b, err := json.Marshal(content)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal project", err.Error())
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
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/branches/%s/databases/%s", data.ProjectID.ValueString(), data.BranchID.ValueString(), data.Name.ValueString())
	request, err := http.NewRequest("PATCH", url, bytes.NewBuffer(b))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create the request", err.Error())
		return
	}
	request.Header.Add("Authorization", "Bearer "+key)
	request.Header.Set("Content-Type", "application/json")
	response, err := r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create database", err.Error())
		return
	} else if response.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError("Failed to create database, response status ", response.Status)
		return
	}
	defer response.Body.Close()
	inner := struct {
		Database databaseResourceJSON `json:"database"`
	}{
		Database: databaseResourceJSON{},
	}
	err = json.NewDecoder(response.Body).Decode(&inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	diags = resp.State.Set(ctx, toDatabaseModel(&inner.Database))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
}

func (r databaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
