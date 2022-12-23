package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type projectResource struct {
	client *resty.Client
}

var _ resource.Resource = projectResource{}
var _ resource.ResourceWithImportState = projectResource{}

func connectionUriResourceAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"connection_uri": schema.StringAttribute{
			Required: true,
		},
	}
}

func (r projectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
			"maintenance_starts_at": schema.StringAttribute{
				MarkdownDescription: "neon host",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "neon host",
				Computed:            true,
			},
			"platform_id": schema.StringAttribute{
				MarkdownDescription: "neon host",
				Computed:            true,
			},
			"region_id": schema.StringAttribute{
				MarkdownDescription: "neon host",
				Computed:            true,
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "neon host",
				Required:            true,
			},
			"engine": schema.StringAttribute{
				MarkdownDescription: "neon host",
				Computed:            true,
				Optional:            true,
				Validators:          []validator.String{stringvalidator.OneOf("k8s-pod", "k8s-neonvm", "docker")},
			},
			"pg_version": schema.Int64Attribute{
				MarkdownDescription: "neon host",
				Optional:            true,
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
			"roles": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: roleResourceAttr(),
				},
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"databases": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: databaseResourceAttr(),
				},
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"branch": schema.SingleNestedAttribute{
				Computed:      true,
				Attributes:    branchResourceAttr(),
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
			},
			"connection_uris": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: connectionUriResourceAttr(),
				},
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
			"endpoints": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: endpointResourceAttr(),
				},
				PlanModifiers: []planmodifier.List{listplanmodifier.UseStateForUnknown()},
			},
		},
		Description:         "",
		MarkdownDescription: "Neon endpoint resource",
		DeprecationMessage:  "",
	}
}

type createProject struct {
	Name        string `json:"name,omitempty"`
	Provisioner string `json:"provisioner,omitempty"`
	RegionID    string `json:"region_id,omitempty"`
	PgVersion   int64  `json:"pg_version,omitempty"`
	//Settings                 *endpointSettingsJSON `json:"settings,omitempty"`
	Autoscaling_limit_min_cu int64 `json:"autoscaling_limit_min_cu,omitempty"`
	Autoscaling_limit_max_cu int64 `json:"autoscaling_limit_max_cu,omitempty"`
}

func (r projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	data := newProjectResourceModel()
	diags := req.Plan.Get(ctx, data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	content := struct {
		Project createProject `json:"project"`
	}{
		Project: createProject{
			Name:                     data.Name.ValueString(),
			Provisioner:              data.Provisioner.ValueString(),
			RegionID:                 data.RegionID.ValueString(),
			PgVersion:                data.PgVersion.ValueInt64(),
			Autoscaling_limit_min_cu: data.AutoscalingLimitMinCu.ValueInt64(),
			Autoscaling_limit_max_cu: data.AutoscalingLimitMaxCu.ValueInt64(),
		},
	}
	response, err := r.client.R().
		SetBody(content).
		Post("/projects")
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create endpoint resource with a status code: %s", response.Status()), "")
		return
	}

	inner := projectResourceJSON{}
	err = json.Unmarshal(response.Body(), &inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}

	if len(inner.Endpoints) > 0 {
		var isEndpointReady bool
		response, _ = r.client.R().AddRetryCondition(func(r *resty.Response, e error) bool {
			if r.IsError() || e != nil {
				return false
			}
			endpoint := struct {
				Endpoint endpointResourceJSON `json:"endpoint"`
			}{}
			err = json.Unmarshal(r.Body(), &endpoint)
			if err != nil {
				return false
			}
			isEndpointReady = endpoint.Endpoint.CurrentState == "init"
			return isEndpointReady
		}).
			Get(fmt.Sprintf("/projects/%s/endpoints/%s", inner.Project.ID, inner.Endpoints[0].Id)) //review
		if response.IsError() {
			resp.Diagnostics.AddError(fmt.Sprintf("Cannot get endpoint status with a status code: %s", response.Status()), "")
			return
		}
	}

	//tflog.Trace(ctx, "created a resource")
	plan, diags := inner.ToProjectResourceModel(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete implements resource.Resource
func (r projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data projectResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	response, _ := r.client.R().Delete(fmt.Sprintf("/projects/%s", data.ID.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), "")
		return
	}
	resp.State.RemoveResource(ctx)
}

// Metadata implements resource.Resource
func (r projectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Read implements resource.Resource
func (r projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	data := newProjectResourceModel()
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	response, _ := r.client.R().Get(fmt.Sprintf("/projects/%s", data.ID.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), "")
		return
	}
	project := &projectResourceJSON{}
	err := json.Unmarshal(response.Body(), &project)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}

	plan, diags := project.ToProjectResourceModel(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

type updateProject struct {
	Name string `json:"name"`
	//Settings                 *endpointSettingsJSON `json:"settings,omitempty"`
	Autoscaling_limit_min_cu int64 `json:"autoscaling_limit_min_cu,omitempty"`
	Autoscaling_limit_max_cu int64 `json:"autoscaling_limit_max_cu,omitempty"`
}

// Update implements resource.Resource
func (r projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	data := newProjectResourceModel()
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var ID types.String
	diags = req.State.GetAttribute(ctx, path.Root("id"), &ID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	u := updateProject{
		Autoscaling_limit_min_cu: data.AutoscalingLimitMinCu.ValueInt64(),
		Autoscaling_limit_max_cu: data.AutoscalingLimitMaxCu.ValueInt64(),
		Name:                     data.Name.ValueString(),
	}
	content := struct {
		Project updateProject `json:"project"`
	}{
		Project: u,
	}
	response, _ := r.client.R().
		SetBody(content).
		Patch(fmt.Sprintf("/projects/%s", ID.ValueString()))
	if response.IsError() {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create endpoint resource with a status code: %s", response.Status()), "")
		return
	}

	project := &projectResourceJSON{}
	err := json.Unmarshal(response.Body(), &project)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	plan, diags := project.ToProjectResourceModel(ctx)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
