package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = endpointDataResource{}

func NewDataResource() datasource.DataSource {
	return endpointDataResource{}
}

type endpointDataResource struct {
	client *http.Client
}

type endpointDataResourceModel struct {
	Host                  types.String `tfsdk:"host"`
	Id                    types.String `tfsdk:"id"`
	ProjectID             types.String `tfsdk:"project_id"`
	BranchID              types.String `tfsdk:"branch_id"`
	AutoscalingLimitMinCu types.Int64  `tfsdk:"autoscaling_limit_min_cu"`
	AutoscalingLimitMaxCu types.Int64  `tfsdk:"autoscaling_limit_max_cu"`
	RegionID              types.String `tfsdk:"region_id"`
	Type                  types.String `tfsdk:"type"`
	CurrentState          types.String `tfsdk:"current_state"`
	PendingState          types.String `tfsdk:"pending_state"`
	Settings              struct {
		Description types.String `tfsdk:"description"`
		PgSettings  types.Map    `tfsdk:"pg_settings"`
	} `tfsdk:"settings"`
	PoolerEnabled      types.Bool   `tfsdk:"pooler_enabled"`
	PoolerMode         types.String `tfsdk:"pooler_mode"`
	Disabled           types.Bool   `tfsdk:"disabled"`
	PasswordlessAccess types.Bool   `tfsdk:"passwordless_access"`
	LastActive         types.String `tfsdk:"last_active"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

func (r endpointDataResource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoint"
}

func (r endpointDataResource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Neon endpoint resource",

		Attributes: map[string]schema.Attribute{
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
				Required:            true,
			},
			"branch_id": schema.StringAttribute{
				MarkdownDescription: "postgres branch",
				Required:            true,
			},
			"autoscaling_limit_min_cu": schema.Int64Attribute{
				MarkdownDescription: "autoscaling limit min",
				Computed:            true,
			},
			"autoscaling_limit_max_cu": schema.Int64Attribute{
				MarkdownDescription: "autoscaling limit max",
				Computed:            true,
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
			"settings": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Required: true,
						},
						"pg_settings": schema.MapNestedAttribute{
							Required: true,
						},
					},
				},
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
		},
	}
}
func (d endpointDataResource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data endpointDataResourceModel
	diags := req.Config.Get(ctx, &data)
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
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/endpoints/%s", data.ProjectID.ValueString(), data.Id.ValueString())
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create the request", err.Error())
		return
	}
	request.Header.Add("Authorization", "Bearer "+key)
	request.Header.Set("Content-Type", "application/json")
	response, err := d.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create endpoint", err.Error())
		return
	}
	defer response.Body.Close()
	endpoint := struct {
		Endpoint endpointDataResourceModel `json:"endpoint"`
	}{}
	err = json.NewDecoder(response.Body).Decode(&endpoint)
	data = endpoint.Endpoint
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	//tflog.Trace(ctx, "created a resource")
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
