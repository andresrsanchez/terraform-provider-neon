package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &projectDataSource{}

type projectDataSource struct {
	client *resty.Client
}

type projectDataModel struct {
	MaintenanceStartsAt types.String `tfsdk:"maintenance_starts_at"`
	ID                  types.String `tfsdk:"id"`
	PlatformID          types.String `tfsdk:"platform_id"`
	RegionID            types.String `tfsdk:"region_id"`
	Name                types.String `tfsdk:"name"`
	Provisioner         types.String `tfsdk:"engine"`
	PgVersion           types.Int64  `tfsdk:"pg_version"`
	LastActive          types.String `tfsdk:"last_active"`
	CreatedAt           types.String `tfsdk:"created_at"`
	UpdatedAt           types.String `tfsdk:"updated_at"`
	Host                types.String `tfsdk:"proxy_host"`
}

type projectDataJSON struct {
	MaintenanceStartsAt string `json:"maintenance_starts_at"`
	ID                  string `json:"id"`
	PlatformID          string `json:"platform_id"`
	RegionID            string `json:"region_id"`
	Name                string `json:"name"`
	Provisioner         string `json:"engine"`
	PgVersion           int64  `json:"pg_version"`
	LastActive          string `json:"last_active"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
	Host                string `json:"proxy_host"`
}

func (in *projectDataJSON) ToProjectDataModel() *projectDataModel {
	return &projectDataModel{
		MaintenanceStartsAt: types.StringValue(in.MaintenanceStartsAt),
		ID:                  types.StringValue(in.ID),
		PlatformID:          types.StringValue(in.PlatformID),
		RegionID:            types.StringValue(in.RegionID),
		Name:                types.StringValue(in.Name),
		Provisioner:         types.StringValue(in.Provisioner),
		PgVersion:           types.Int64Value(in.PgVersion),
		LastActive:          types.StringValue(in.LastActive),
		CreatedAt:           types.StringValue(in.CreatedAt),
		UpdatedAt:           types.StringValue(in.UpdatedAt),
		Host:                types.StringValue(in.Host),
	}
}

// Metadata implements datasource.DataSource
func (*projectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Read implements datasource.DataSource
func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data projectDataModel
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	response, _ := d.client.R().Get(fmt.Sprintf("/projects/%s", data.ID.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), "")
		return
	}
	project := struct {
		Project projectDataJSON `json:"project"`
	}{}
	err := json.Unmarshal(response.Body(), &project)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}

	plan := project.Project.ToProjectDataModel()
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Schema implements datasource.DataSource
func (*projectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"autoscaling_limit_min_cu": schema.Int64Attribute{
				MarkdownDescription: "autoscaling limit min",
				Computed:            true,
			},
			"autoscaling_limit_max_cu": schema.Int64Attribute{
				MarkdownDescription: "autoscaling limit max",
				Computed:            true,
			},
			"maintenance_starts_at": schema.StringAttribute{
				MarkdownDescription: "neon host",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "neon host",
				Required:            true,
			},
			"platform_id": schema.StringAttribute{
				MarkdownDescription: "neon host",
				Computed:            true,
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "neon host",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "neon host",
				Required:            true,
			},
			"engine": schema.StringAttribute{
				MarkdownDescription: "neon host",
				Computed:            true,
				Validators:          []validator.String{stringvalidator.OneOf("k8s-pod", "k8s-neonvm", "docker")},
			},
			"pg_version": schema.Int64Attribute{
				MarkdownDescription: "neon host",
				Computed:            true,
				Validators:          []validator.Int64{int64validator.Any(int64validator.OneOf(14), int64validator.OneOf(15))},
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
		},
	}
}
