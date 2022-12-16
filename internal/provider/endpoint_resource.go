package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ resource.Resource = endpointResource{}
var _ resource.ResourceWithImportState = endpointResource{}

type endpointResource struct {
	client *resty.Client
}

func (r endpointResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_endpoint"
}

func (r endpointResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Neon endpoint resource",
		Attributes:          endpointResourceAttr(),
	}
}

func endpointResourceAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
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
			Computed:            true,
		},
		"branch_id": schema.StringAttribute{
			MarkdownDescription: "postgres branch",
			Computed:            true,
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
		/*"settings": schema.ObjectAttribute{
			AttributeTypes: map[string]attr.Type{
				"description": types.StringType,
				"pg_settings": types.ListType{
					ElemType: types.MapType{
						ElemType: types.StringType,
					},
				},
			},
			Computed: true,
		},*/
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
	}
}

type createEndpoint struct {
	Branch_id                string                `json:"branch_id"`
	Region_id                string                `json:"region_id,omitempty"`
	Type                     string                `json:"type"`
	Settings                 *endpointSettingsJSON `json:"settings,omitempty"`
	Autoscaling_limit_min_cu int64                 `json:"autoscaling_limit_min_cu,omitempty"`
	Autoscaling_limit_max_cu int64                 `json:"autoscaling_limit_max_cu,omitempty"`
	Pooler_enabled           bool                  `json:"pooler_enabled,omitempty"`
	Pooler_mode              string                `json:"pooler_mode,omitempty"`
	Disabled                 bool                  `json:"disabled,omitempty"`
	Passwordless_access      bool                  `json:"passwordless_access,omitempty"`
}

func (r endpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	fmt.Println("hi2")
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
	/*if data.Settings != nil {
		content.Settings = data.Settings.ToEndpointSettingsJSON()
	}*/
	response, err := r.client.R().
		SetBody(content).
		Post(fmt.Sprintf("/projects/%s/endpoints", data.ProjectID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create endpoint resource with a status code: %s", response.Status()), err.Error())
		return
	}
	endpoint := struct {
		Endpoint endpointResourceJSON `json:"endpoint"`
	}{}
	err = json.Unmarshal(response.Body(), &endpoint)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	data = *endpoint.Endpoint.ToEndpointResourceModel()

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
	response, err := r.client.R().Get(fmt.Sprintf("/projects/%s/endpoints/%s", data.ProjectID.ValueString(), data.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), err.Error())
		return
	}

	endpoint := struct {
		Endpoint endpointResourceJSON `json:"endpoint"`
	}{}
	err = json.Unmarshal(response.Body(), &endpoint)
	data = *endpoint.Endpoint.ToEndpointResourceModel()
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

type updateEndpoint struct {
	Branch_id                string                `json:"branch_id"`
	Settings                 *endpointSettingsJSON `json:"settings,omitempty"`
	Autoscaling_limit_min_cu int64                 `json:"autoscaling_limit_min_cu,omitempty"`
	Autoscaling_limit_max_cu int64                 `json:"autoscaling_limit_max_cu,omitempty"`
	Pooler_enabled           bool                  `json:"pooler_enabled,omitempty"`
	Pooler_mode              string                `json:"pooler_mode,omitempty"`
	Disabled                 bool                  `json:"disabled,omitempty"`
	Passwordless_access      bool                  `json:"passwordless_access,omitempty"`
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
	/*if !data.Settings.Description.IsUnknown() {
		content.Settings = data.Settings.ToEndpointSettingsJSON()
	}*/
	response, err := r.client.R().
		SetBody(content).
		Patch(fmt.Sprintf("/projects/%s/endpoints", data.ProjectID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create endpoint resource with a status code: %s", response.Status()), err.Error())
		return
	}

	endpoint := struct {
		Endpoint endpointResourceJSON `json:"endpoint"`
	}{}
	err = json.Unmarshal(response.Body(), &endpoint)
	data = *endpoint.Endpoint.ToEndpointResourceModel()
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

	response, err := r.client.R().Delete(fmt.Sprintf("/projects/%s/endpoints/%s", data.ProjectID.ValueString(), data.Id.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r endpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
