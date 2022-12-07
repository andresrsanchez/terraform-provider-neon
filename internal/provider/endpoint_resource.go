package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = endpointResource{}
var _ resource.ResourceWithImportState = endpointResource{}

func NewEndpointResource() resource.Resource {
	return endpointResource{
		client: &http.Client{},
	}
}

type endpointResource struct {
	client *http.Client
}

type endpointResourceModel struct {
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
		PgSettings  types.List   `tfsdk:"pg_settings"`
	} `tfsdk:"settings"`
	PoolerEnabled      types.Bool   `tfsdk:"pooler_enabled"`
	PoolerMode         types.String `tfsdk:"pooler_mode"`
	Disabled           types.Bool   `tfsdk:"disabled"`
	PasswordlessAccess types.Bool   `tfsdk:"passwordless_access"`
	LastActive         types.String `tfsdk:"last_active"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

func (r endpointResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoint"
}

func (r endpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Optional:            true,
			},
			"autoscaling_limit_max_cu": schema.Int64Attribute{
				MarkdownDescription: "autoscaling limit max",
				Computed:            true,
				Optional:            true,
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "region id",
				Computed:            true,
				Optional:            true,
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
			"settings": schema.ObjectAttribute{
				AttributeTypes: map[string]attr.Type{
					"description": types.StringType,
					"pg_settings": types.ListType{
						ElemType: types.MapType{
							ElemType: types.StringType,
						},
					},
				},
				Computed: true,
				Optional: true,
			},
			"pooler_enabled": schema.BoolAttribute{
				MarkdownDescription: "pooler enabled",
				Computed:            true,
				Optional:            true,
			},
			"pooler_mode": schema.StringAttribute{
				MarkdownDescription: "pooler mode",
				Computed:            true,
				Optional:            true,
				Validators:          []validator.String{stringvalidator.OneOf("transaction")},
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "disabled",
				Computed:            true,
				Optional:            true,
			},
			"passwordless_access": schema.BoolAttribute{
				MarkdownDescription: "passwordless access",
				Computed:            true,
				Optional:            true,
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

/*func (r endpointResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}*/

type endpointSettings struct {
	Description string            `json:"description"`
	Pg_settings map[string]string `json:"pg_settings"`
}
type createEndpoint struct {
	Branch_id                string           `json:"branch_id"`
	Region_id                string           `json:"region_id,omitempty"`
	Type                     string           `json:"type"`
	Settings                 endpointSettings `json:"settings,omitempty"`
	Autoscaling_limit_min_cu int64            `json:"autoscaling_limit_min_cu,omitempty"`
	Autoscaling_limit_max_cu int64            `json:"autoscaling_limit_max_cu,omitempty"`
	Pooler_enabled           bool             `json:"pooler_enabled,omitempty"`
	Pooler_mode              string           `json:"pooler_mode,omitempty"`
	Disabled                 bool             `json:"disabled,omitempty"`
	Passwordless_access      bool             `json:"passwordless_access,omitempty"`
}

func (r endpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data endpointResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	content := createEndpoint{
		Branch_id:                data.BranchID.ValueString(),
		Region_id:                data.RegionID.ValueString(),
		Type:                     data.Type.ValueString(),
		Autoscaling_limit_min_cu: data.AutoscalingLimitMinCu.ValueInt64(),
		Autoscaling_limit_max_cu: data.AutoscalingLimitMaxCu.ValueInt64(),
		Pooler_enabled:           data.PoolerEnabled.ValueBool(),
		Pooler_mode:              data.PoolerMode.ValueString(),
		Disabled:                 data.Disabled.ValueBool(),
		Passwordless_access:      data.PasswordlessAccess.ValueBool(),
	}
	if !data.Settings.Description.IsUnknown() {
		s := make(map[string]string)
		for _, v := range data.Settings.PgSettings.Elements() {
			pg_settings := v.(types.Map)
			for k, setting := range pg_settings.Elements() {
				s[k] = setting.String()
			}
		}
		content.Settings = endpointSettings{
			data.Settings.Description.ValueString(),
			s,
		}
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
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/endpoints", data.ProjectID.ValueString())
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create the request", err.Error())
		return
	}
	request.Header.Add("Authorization", "Bearer "+key)
	request.Header.Set("Content-Type", "application/json")
	response, err := r.client.Do(request)
	if err != nil || response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Failed to create endpoint", err.Error())
		return
	}
	defer response.Body.Close()
	endpoint := struct {
		Endpoint endpointResourceModel `json:"endpoint"`
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
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
}

func (r endpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data endpointResourceModel
	diags := req.State.Get(ctx, &data)
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
	response, err := r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create endpoint", err.Error())
		return
	}
	defer response.Body.Close()
	endpoint := struct {
		Endpoint endpointResourceModel `json:"endpoint"`
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
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

type updateEndpoint struct {
	Branch_id                string           `json:"branch_id"`
	Settings                 endpointSettings `json:"settings,omitempty"`
	Autoscaling_limit_min_cu int64            `json:"autoscaling_limit_min_cu,omitempty"`
	Autoscaling_limit_max_cu int64            `json:"autoscaling_limit_max_cu,omitempty"`
	Pooler_enabled           bool             `json:"pooler_enabled,omitempty"`
	Pooler_mode              string           `json:"pooler_mode,omitempty"`
	Disabled                 bool             `json:"disabled,omitempty"`
	Passwordless_access      bool             `json:"passwordless_access,omitempty"`
}

func (r endpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data endpointResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	content := updateEndpoint{
		Branch_id:                data.BranchID.ValueString(),
		Autoscaling_limit_min_cu: data.AutoscalingLimitMinCu.ValueInt64(),
		Autoscaling_limit_max_cu: data.AutoscalingLimitMaxCu.ValueInt64(),
		Pooler_enabled:           data.PoolerEnabled.ValueBool(),
		Pooler_mode:              data.PoolerMode.ValueString(),
		Disabled:                 data.Disabled.ValueBool(),
		Passwordless_access:      data.PasswordlessAccess.ValueBool(),
	}
	if !data.Settings.Description.IsUnknown() {
		s := make(map[string]string)
		for _, v := range data.Settings.PgSettings.Elements() {
			pg_settings := v.(types.Map)
			for k, setting := range pg_settings.Elements() {
				s[k] = setting.String()
			}
		}
		content.Settings = endpointSettings{
			data.Settings.Description.ValueString(),
			s,
		}
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
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/endpoints", data.ProjectID.ValueString())
	request, err := http.NewRequest("PATCH", url, bytes.NewBuffer(b))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create the request", err.Error())
		return
	}
	request.Header.Add("Authorization", "Bearer "+key)
	request.Header.Set("Content-Type", "application/json")
	response, err := r.client.Do(request)
	if err != nil || response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Failed to update endpoint", err.Error())
		return
	}
	defer response.Body.Close()
	endpoint := struct {
		Endpoint endpointResourceModel `json:"endpoint"`
	}{}
	err = json.NewDecoder(response.Body).Decode(&endpoint)
	data = endpoint.Endpoint
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
}

func (r endpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data endpointResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s/endpoints", data.ProjectID.ValueString())
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
	_, err = r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create endpoint", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r endpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
