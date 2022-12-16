package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type projectConnUrisJSON struct {
	ConnectionURI string `json:"connection_uri"`
}

type endpointSettingsJSON struct {
	PgSettings map[string]string `json:"pg_settings"`
}
type innerProjectResourceJSON struct {
	MaintenanceStartsAt   string                `json:"maintenance_starts_at"`
	ID                    string                `json:"id"`
	PlatformID            string                `json:"platform_id"`
	RegionID              string                `json:"region_id"`
	Name                  string                `json:"name"`
	Provisioner           string                `json:"provisioner"`
	Settings              *endpointSettingsJSON `json:"settings"`
	PgVersion             int64                 `json:"pg_version"`
	AutoscalingLimitMinCu int64                 `json:"autoscaling_limit_min_cu"`
	AutoscalingLimitMaxCu int64                 `json:"autoscaling_limit_max_cu"`
	LastActive            string                `json:"last_active"`
	CreatedAt             string                `json:"created_at"`
	UpdatedAt             string                `json:"updated_at"`
}
type projectResourceJSON struct {
	Project        innerProjectResourceJSON `json:"project"`
	ConnectionUris []projectConnUrisJSON    `json:"connection_uris"`
	Roles          []roleResourceJSON       `json:"roles"`
	Databases      []databaseResourceJSON   `json:"databases"`
	Branch         *branchResourceJSON      `json:"branch"`
	Endpoints      []endpointResourceJSON   `json:"endpoints"`
}

type projectConnUrisModel struct {
	ConnectionURI types.String `tfsdk:"connection_uri"`
}

type endpointSettingsModel struct {
	//Description types.String `tfsdk:"description"`
	PgSettings types.Map `tfsdk:"pg_settings"`
}
type projectResourceModel struct {
	MaintenanceStartsAt types.String `tfsdk:"maintenance_starts_at"`
	ID                  types.String `tfsdk:"id"`
	PlatformID          types.String `tfsdk:"platform_id"`
	RegionID            types.String `tfsdk:"region_id"`
	Name                types.String `tfsdk:"name"`
	Provisioner         types.String `tfsdk:"engine"`
	//Settings              types.Object `tfsdk:"settings"`
	PgVersion             types.Int64  `tfsdk:"pg_version"`
	AutoscalingLimitMinCu types.Int64  `tfsdk:"autoscaling_limit_min_cu"`
	AutoscalingLimitMaxCu types.Int64  `tfsdk:"autoscaling_limit_max_cu"`
	LastActive            types.String `tfsdk:"last_active"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
	ConnectionUris        types.List   `tfsdk:"connection_uris"`
	Roles                 types.List   `tfsdk:"roles"`
	Databases             types.List   `tfsdk:"databases"`
	Branch                types.Object `tfsdk:"branch"`
	Endpoints             types.List   `tfsdk:"endpoints"`
}

func newProjectResourceModel() *projectResourceModel {
	return &projectResourceModel{
		//Settings:       types.ObjectNull(typeFromAttrs(defaultSettingsResourceAttr())),
		ConnectionUris: types.ListNull(types.ObjectType{AttrTypes: typeFromAttrs(connectionUriResourceAttr())}),
		Roles:          types.ListNull(types.ObjectType{AttrTypes: typeFromAttrs(roleResourceAttr())}),
		Databases:      types.ListNull(types.ObjectType{AttrTypes: typeFromAttrs(databaseResourceAttr())}),
		Branch:         types.ObjectNull(typeFromAttrs(branchResourceAttr())),
		Endpoints:      types.ListNull(types.ObjectType{AttrTypes: typeFromAttrs(endpointResourceAttr())}),
	}
}

func (p *projectResourceJSON) ToProjectResourceModel() (*projectResourceModel, diag.Diagnostics) {
	m := newProjectResourceModel()
	m.MaintenanceStartsAt = types.StringValue(p.Project.MaintenanceStartsAt)
	m.ID = types.StringValue(p.Project.ID)
	m.PlatformID = types.StringValue(p.Project.PlatformID)
	m.RegionID = types.StringValue(p.Project.RegionID)
	m.Name = types.StringValue(p.Project.Name)
	m.Provisioner = types.StringValue(p.Project.Provisioner)
	m.PgVersion = types.Int64Value(p.Project.PgVersion)
	m.AutoscalingLimitMinCu = types.Int64Value(p.Project.AutoscalingLimitMinCu)
	m.AutoscalingLimitMaxCu = types.Int64Value(p.Project.AutoscalingLimitMaxCu)
	m.LastActive = types.StringValue(p.Project.LastActive)
	m.CreatedAt = types.StringValue(p.Project.CreatedAt)
	m.UpdatedAt = types.StringValue(p.Project.UpdatedAt)

	/*if p.Settings != nil {
		pg, _ := types.MapValueFrom(context.TODO(), types.StringType, p.Settings.PgSettings)
		settings := endpointSettingsModel{
			PgSettings: pg,
		}
		aux, _ := types.ObjectValueFrom(context.TODO(), typeFromAttrs(defaultSettingsResourceAttr()), settings)
		m.Settings = aux
	}*/
	if p.Branch != nil {
		branchModel, diags := p.Branch.ToBranchResourceModel()
		if diags.HasError() {
			return nil, diags
		}
		branchObj, diags := branchModel.ToBranchResourceObject()
		if diags.HasError() {
			return nil, diags
		}
		m.Branch = branchObj
	}
	if len(p.ConnectionUris) != 0 {
		c := []projectConnUrisModel{}
		for _, v := range p.ConnectionUris {
			c = append(c, projectConnUrisModel{
				ConnectionURI: types.StringValue(v.ConnectionURI),
			})
		}
		aux, diags := types.ListValueFrom(context.TODO(), types.ObjectType{AttrTypes: typeFromAttrs(connectionUriResourceAttr())}, c)
		if diags.HasError() {
			return nil, diags
		}
		m.ConnectionUris = aux
	}
	if len(p.Roles) != 0 {
		r := []types.Object{}
		for _, v := range p.Roles {
			aux, diags := toRoleModel(&v)
			if diags.HasError() {
				return nil, diags
			}
			r = append(r, aux)
		}
		aux, diags := types.ListValueFrom(context.TODO(), types.ObjectType{AttrTypes: typeFromAttrs(roleResourceAttr())}, r)
		if diags.HasError() {
			return nil, diags
		}
		m.Roles = aux
	}
	if len(p.Databases) != 0 {
		d := []types.Object{}
		for _, v := range p.Databases {
			aux, diags := toDatabaseModel(&v, p.Project.ID)
			if diags.HasError() {
				return nil, diags
			}
			d = append(d, aux)
		}
		aux, diags := types.ListValueFrom(context.TODO(), types.ObjectType{AttrTypes: typeFromAttrs(databaseResourceAttr())}, d)
		if diags.HasError() {
			return nil, diags
		}
		m.Databases = aux
	}
	if len(p.Endpoints) != 0 {
		e := []endpointResourceModel{}
		for _, v := range p.Endpoints {
			ee := *v.ToEndpointResourceModel()
			e = append(e, ee)
		}
		aux, diags := types.ListValueFrom(context.TODO(), types.ObjectType{AttrTypes: typeFromAttrs(endpointResourceAttr())}, e)
		if diags.HasError() {
			return nil, diags
		}
		m.Endpoints = aux
	}
	return m, nil
}
func typeFromAttrs(in map[string]schema.Attribute) map[string]attr.Type {
	out := map[string]attr.Type{}
	for k, v := range in {
		out[k] = v.GetType()
	}
	return out
}
func (m *projectResourceModel) ToProjectResourceJSON() (*projectResourceJSON, diag.Diagnostics) {
	p := &projectResourceJSON{
		Project: innerProjectResourceJSON{
			MaintenanceStartsAt:   m.MaintenanceStartsAt.ValueString(),
			ID:                    m.ID.ValueString(),
			PlatformID:            m.PlatformID.ValueString(),
			RegionID:              m.RegionID.ValueString(),
			Name:                  m.Name.ValueString(),
			Provisioner:           m.Provisioner.ValueString(),
			PgVersion:             m.PgVersion.ValueInt64(),
			AutoscalingLimitMinCu: m.AutoscalingLimitMinCu.ValueInt64(),
			AutoscalingLimitMaxCu: m.AutoscalingLimitMaxCu.ValueInt64(),
			LastActive:            m.LastActive.ValueString(),
			CreatedAt:             m.CreatedAt.ValueString(),
			UpdatedAt:             m.UpdatedAt.ValueString(),
		},
	}
	/*if !m.Settings.IsNull() {
		v := endpointSettingsModel{}
		m.Settings.As(context.TODO(), &v, basetypes.ObjectAsOptions{})
		p.Settings = v.ToEndpointSettingsJSON()
	}*/
	if !m.Branch.IsNull() {
		aux, diags := toBranchJSON(&m.Branch)
		if diags.HasError() {
			return nil, diags
		}
		p.Branch = aux
	}
	if !m.ConnectionUris.IsNull() {
		p.ConnectionUris = []projectConnUrisJSON{}
		for _, vv := range m.ConnectionUris.Elements() {
			v := projectConnUrisModel{}
			diags := vv.(types.Object).As(context.TODO(), &v, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return nil, diags
			}
			p.ConnectionUris = append(p.ConnectionUris, projectConnUrisJSON{
				ConnectionURI: v.ConnectionURI.ValueString(),
			})
		}
	}
	if !m.Roles.IsNull() {
		p.Roles = []roleResourceJSON{}
		for _, vv := range m.Roles.Elements() {
			roleObj := vv.(types.Object)
			role, diags := toRoleJSON(&roleObj)
			if diags.HasError() {
				return nil, diags
			}
			p.Roles = append(p.Roles, *role)
		}
	}
	if !m.Databases.IsNull() {
		p.Databases = []databaseResourceJSON{}
		for _, vv := range m.Databases.Elements() {
			databaseObj := vv.(types.Object)
			aux, diags := toDatabaseJSON(&databaseObj)
			if diags.HasError() {
				return nil, diags
			}
			p.Databases = append(p.Databases, *aux)
		}
	}
	if m.Endpoints.IsNull() {
		p.Endpoints = []endpointResourceJSON{}
		for _, vv := range m.Endpoints.Elements() {
			v := endpointResourceModel{}
			diags := vv.(types.Object).As(context.TODO(), &v, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				return nil, diags
			}
			p.Endpoints = append(p.Endpoints, *v.ToEndpointResourceJSON())
		}
	}
	return p, nil
}
