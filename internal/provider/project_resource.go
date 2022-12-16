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

func defaultSettingsResourceAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		/*"description": schema.StringAttribute{
			Required: true,
		},*/
		"pg_settings": schema.MapAttribute{
			Computed:    true,
			Optional:    true,
			ElementType: types.StringType,
		},
	}
}

//migrate to projectattrs()
func (r projectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			/*"settings": schema.SingleNestedAttribute{
				Attributes: defaultSettingsResourceAttr(),
				Optional:   true,
				Computed:   true,
			},*/
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
			},
			"databases": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: databaseResourceAttr(),
				},
			},
			"branch": schema.SingleNestedAttribute{
				Computed:   true,
				Attributes: branchResourceAttr(),
			},
			"connection_uris": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: connectionUriResourceAttr(),
				},
			},
			"endpoints": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: endpointResourceAttr(),
				},
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
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create endpoint resource with a status code: %s", response.Status()), err.Error())
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
		response, err = r.client.R().AddRetryCondition(func(r *resty.Response, err error) bool {
			if err != nil || r.StatusCode() > 399 {
				return true
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
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Cannot get endpoint status with a status code: %s", response.Status()), err.Error())
			return
		}
	}

	//tflog.Trace(ctx, "created a resource")
	plan, diags := inner.ToProjectResourceModel()
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

	response, err := r.client.R().Delete(fmt.Sprintf("/projects/%s", data.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), err.Error())
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
	fmt.Println("hola1")
	data := newProjectResourceModel()
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	response, err := r.client.R().Get(fmt.Sprintf("/projects/%s", data.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to delete branch resource with a status code: %s", response.Status()), err.Error())
		return
	}
	project := &projectResourceJSON{}
	err = json.Unmarshal(response.Body(), &project)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	plan, diags := project.ToProjectResourceModel()
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

type updateProject struct {
	Name string `tfsdk:"name"`
	//Settings                 *endpointSettingsJSON `json:"settings,omitempty"`
	Autoscaling_limit_min_cu int64 `json:"autoscaling_limit_min_cu,omitempty"`
	Autoscaling_limit_max_cu int64 `json:"autoscaling_limit_max_cu,omitempty"`
}

// Update implements resource.Resource
func (r projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	fmt.Println("hola2")
	data := newProjectResourceModel()

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	u := updateProject{
		Autoscaling_limit_min_cu: data.AutoscalingLimitMinCu.ValueInt64(),
		Autoscaling_limit_max_cu: data.AutoscalingLimitMaxCu.ValueInt64(),
		Name:                     data.Name.ValueString(),
	}
	/*if !data.Settings.IsNull() {
		v := endpointSettingsModel{}
		data.Settings.As(context.TODO(), &v, basetypes.ObjectAsOptions{})
		u.Settings = v.ToEndpointSettingsJSON()
	}*/
	content := struct {
		Project updateProject `json:"project"`
	}{
		Project: u,
	}
	response, err := r.client.R().
		SetBody(content).
		Patch(fmt.Sprintf("/projects/%s", data.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Failed to create endpoint resource with a status code: %s", response.Status()), err.Error())
		return
	}

	project := &projectResourceJSON{}
	err = json.Unmarshal(response.Body(), &project)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	plan, diags := project.ToProjectResourceModel()
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	fmt.Println("hola3")
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
