package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type projectResource struct {
	client *http.Client
}

var _ resource.Resource = projectResource{}
var _ resource.ResourceWithImportState = projectResource{}

func NewProjectResource() resource.Resource {
	return projectResource{
		client: &http.Client{},
	}
}

func roleResourceAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"branch_id": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
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

func databaseResourceAttr() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed: true,
		},
		"branch_id": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"owner_name": schema.StringAttribute{
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
			"provisioner": schema.StringAttribute{
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

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
	c := createProject{
		Name:                     data.Name.ValueString(),
		Provisioner:              data.Provisioner.ValueString(),
		RegionID:                 data.RegionID.ValueString(),
		PgVersion:                data.PgVersion.ValueInt64(),
		Autoscaling_limit_min_cu: data.AutoscalingLimitMinCu.ValueInt64(),
		Autoscaling_limit_max_cu: data.AutoscalingLimitMaxCu.ValueInt64(),
	}
	/*if !data.Settings.IsNull() {
		v := endpointSettingsModel{}
		data.Settings.As(context.TODO(), &v, types.ObjectAsOptions{})
		c.Settings = v.ToEndpointSettingsJSON()
	}*/
	content := struct {
		Project createProject `json:"project"`
	}{
		Project: c,
	}
	b, err := json.Marshal(content)
	if err != nil {
		resp.Diagnostics.AddError("Failed to marshal project", err.Error())
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
	request, err := http.NewRequest("POST", "https://console.neon.tech/api/v2/projects", bytes.NewBuffer(b))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create the request", err.Error())
		return
	}
	request.Header.Add("Authorization", "Bearer "+key)
	request.Header.Set("Content-Type", "application/json")
	response, err := r.client.Do(request)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create project", err.Error())
		return
	} else if response.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError("Failed to create project, response status ", response.Status)
		return
	}
	defer response.Body.Close()
	inner := projectResourceJSON{}
	err = json.NewDecoder(response.Body).Decode(&inner)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	//tflog.Trace(ctx, "created a resource")
	plan := inner.ToProjectResourceModel()
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
}

// Delete implements resource.Resource
func (r projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data projectResourceModel
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)

	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s", data.ID.ValueString())
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
		resp.Diagnostics.AddError("Failed to delete project", err.Error())
		return
	} else if response.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError("Failed to delete project, response status ", response.Status)
		return
	}
	resp.State.RemoveResource(ctx)
	if resp.Diagnostics.HasError() {
		return
	}
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
	key, ok := os.LookupEnv("NEON_API_KEY")
	if !ok {
		resp.Diagnostics.AddError(
			"Unable to find token",
			"token cannot be an empty string",
		)
		return
	}
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s", data.ID.ValueString())
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create the request", err.Error())
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
	project := &projectResourceJSON{}
	err = json.NewDecoder(response.Body).Decode(project)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	//tflog.Trace(ctx, "created a resource")
	diags = resp.State.Set(ctx, project.ToProjectResourceModel())
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

type updateProject struct {
	Name string `tfsdk:"name"`
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
	u := updateProject{
		Autoscaling_limit_min_cu: data.AutoscalingLimitMinCu.ValueInt64(),
		Autoscaling_limit_max_cu: data.AutoscalingLimitMaxCu.ValueInt64(),
		Name:                     data.Name.ValueString(),
	}
	/*if !data.Settings.IsNull() {
		v := endpointSettingsModel{}
		data.Settings.As(context.TODO(), &v, types.ObjectAsOptions{})
		u.Settings = v.ToEndpointSettingsJSON()
	}*/
	content := struct {
		Project updateProject `json:"project"`
	}{
		Project: u,
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
	url := fmt.Sprintf("https://console.neon.tech/api/v2/projects/%s", data.ID.ValueString())
	request, err := http.NewRequest("PATCH", url, bytes.NewBuffer(b))
	if err != nil {
		resp.Diagnostics.AddError("Failed to create the request", err.Error())
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
	project := &projectResourceJSON{}
	err = json.NewDecoder(response.Body).Decode(project)
	if err != nil {
		resp.Diagnostics.AddError("Failed to unmarshal response", err.Error())
		return
	}
	//tflog.Trace(ctx, "created a resource")
	diags = resp.State.Set(ctx, project.ToProjectResourceModel())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		//what should i do?
		return
	}
}

func (r projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
