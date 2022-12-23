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
var _ datasource.DataSource = &databaseDataSource{}

type databaseDataSource struct {
	client *resty.Client
}

type databaseDataModel struct {
	ID        types.Int64  `tfsdk:"id"`
	BranchID  types.String `tfsdk:"branch_id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	OwnerName types.String `tfsdk:"owner_name"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

type databaseDataJSON struct {
	ID        int64  `json:"id"`
	BranchID  string `json:"branch_id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	OwnerName string `json:"owner_name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (in *databaseDataJSON) ToDatabaseDataModel() *databaseDataModel {
	return &databaseDataModel{
		ID:        types.Int64Value(in.ID),
		BranchID:  types.StringValue(in.BranchID),
		ProjectID: types.StringValue(in.ProjectID),
		Name:      types.StringValue(in.Name),
		OwnerName: types.StringValue(in.OwnerName),
		CreatedAt: types.StringValue(in.CreatedAt),
		UpdatedAt: types.StringValue(in.UpdatedAt),
	}
}

// Metadata implements datasource.DataSource
func (*databaseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

// Read implements datasource.DataSource
func (d *databaseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data databaseDataModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, _ := d.client.R().Get(fmt.Sprintf("/projects/%s/branches/%s/databases/%s", data.ProjectID.ValueString(), data.BranchID.ValueString(), data.Name.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), "")
		return
	}
	database := struct {
		Database databaseDataJSON `json:"database"`
	}{}
	err := json.Unmarshal(response.Body(), &database)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}

	plan := database.Database.ToDatabaseDataModel()
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Schema implements datasource.DataSource
func (*databaseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
		},
	}
}
