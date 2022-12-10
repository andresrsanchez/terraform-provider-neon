package provider

import (
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type projectResource struct {
	client *http.Client
}
type roleResourceJSON struct {
	BranchID  string `json:"branch_id"`
	Name      string `json:"name"`
	Password  string `json:"password"`
	Protected bool   `json:"protected"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
type databaseResourceJSON struct {
	ID        int64  `json:"id"`
	BranchID  string `json:"branch_id"`
	Name      string `json:"name"`
	OwnerName string `json:"owner_name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
type branchResourceJSON struct {
	ID           string `json:"id"`
	ProjectID    string `json:"project_id"`
	ParentID     string `json:"parent_id"`
	ParentLsn    string `json:"parent_lsn"`
	Name         string `json:"name"`
	CurrentState string `json:"current_state"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}
type projectConnUrisJSON struct {
	ConnectionURI string `json:"connection_uri"`
}

type endpointSettingsJSON struct {
	Description string            `json:"description"`
	PgSettings  map[string]string `json:"pg_settings"`
}
type innerProjectJSON struct {
	MaintenanceStartsAt   string                `json:"maintenance_starts_at"`
	ID                    string                `json:"id"`
	PlatformID            string                `json:"platform_id"`
	RegionID              string                `json:"region_id"`
	Name                  string                `json:"name"`
	Provisioner           string                `json:"provisioner"`
	Settings              *endpointSettingsJSON `json:"default_endpoint_settings"`
	PgVersion             int64                 `json:"pg_version"`
	AutoscalingLimitMinCu int64                 `json:"autoscaling_limit_min_cu"`
	AutoscalingLimitMaxCu int64                 `json:"autoscaling_limit_max_cu"`
	LastActive            string                `json:"last_active"`
	CreatedAt             string                `json:"created_at"`
	UpdatedAt             string                `json:"updated_at"`
}
type projectResourceJSON struct {
	Project        innerProjectJSON       `json:"project"`
	ConnectionUris []projectConnUrisJSON  `json:"connection_uris"`
	Roles          []roleResourceJSON     `json:"roles"`
	Databases      []databaseResourceJSON `json:"databases"`
	Branch         *branchResourceJSON    `json:"branch"`
	Endpoints      []endpointResourceJSON `json:"endpoints"`
}

type roleResourceModel struct {
	BranchID  types.String `tfsdk:"branch_id"`
	Name      types.String `tfsdk:"name"`
	Password  types.String `tfsdk:"password"`
	Protected types.Bool   `tfsdk:"protected"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}
type databaseResourceModel struct {
	ID        types.Int64  `tfsdk:"id"`
	BranchID  types.String `tfsdk:"branch_id"`
	Name      types.String `tfsdk:"name"`
	OwnerName types.String `tfsdk:"owner_name"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}
type branchResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ProjectID    types.String `tfsdk:"project_id"`
	ParentID     types.String `tfsdk:"parent_id"`
	ParentLsn    types.String `tfsdk:"parent_lsn"`
	Name         types.String `tfsdk:"name"`
	CurrentState types.String `tfsdk:"current_state"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}
type projectConnUris struct {
	ConnectionURI types.String `tfsdk:"connection_uri"`
}

type endpointSettingsModel struct {
	Description types.String `tfsdk:"description"`
	PgSettings  types.Map    `tfsdk:"pg_settings"`
}
type innerProjectModel struct {
	MaintenanceStartsAt   types.String           `tfsdk:"maintenance_starts_at"`
	ID                    types.String           `tfsdk:"id"`
	PlatformID            types.String           `tfsdk:"platform_id"`
	RegionID              types.String           `tfsdk:"region_id"`
	Name                  types.String           `tfsdk:"name"`
	Provisioner           types.String           `tfsdk:"provisioner"`
	Settings              *endpointSettingsModel `tfsdk:"default_endpoint_settings"`
	PgVersion             types.Int64            `tfsdk:"pg_version"`
	AutoscalingLimitMinCu types.Int64            `tfsdk:"autoscaling_limit_min_cu"`
	AutoscalingLimitMaxCu types.Int64            `tfsdk:"autoscaling_limit_max_cu"`
	LastActive            types.String           `tfsdk:"last_active"`
	CreatedAt             types.String           `tfsdk:"created_at"`
	UpdatedAt             types.String           `tfsdk:"updated_at"`
}
type projectResourceModel struct {
	Project        innerProjectModel       `tfsdk:"project"`
	ConnectionUris []projectConnUris       `tfsdk:"connection_uris"`
	Roles          []roleResourceModel     `tfsdk:"roles"`
	Databases      []databaseResourceModel `tfsdk:"databases"`
	Branch         *branchResourceModel    `tfsdk:"branch"`
	Endpoints      []endpointResourceModel `tfsdk:"endpoints"`
}

func (p *projectResourceJSON) ToProjectResourceModel() *projectResourceModel {
	m := &projectResourceModel{
		Project: innerProjectModel{
			MaintenanceStartsAt:   types.StringValue(p.Project.MaintenanceStartsAt),
			ID:                    types.StringValue(p.Project.ID),
			PlatformID:            types.StringValue(p.Project.PlatformID),
			RegionID:              types.StringValue(p.Project.RegionID),
			Name:                  types.StringValue(p.Project.Name),
			Provisioner:           types.StringValue(p.Project.Provisioner),
			PgVersion:             types.Int64Value(p.Project.PgVersion),
			AutoscalingLimitMinCu: types.Int64Value(p.Project.AutoscalingLimitMinCu),
			AutoscalingLimitMaxCu: types.Int64Value(p.Project.AutoscalingLimitMaxCu),
			LastActive:            types.StringValue(p.Project.LastActive),
			CreatedAt:             types.StringValue(p.Project.CreatedAt),
			UpdatedAt:             types.StringValue(p.Project.UpdatedAt),
		},
	}
	if p.Project.Settings != nil {
		m.Project.Settings = p.Project.Settings.ToEndpointSettingsModel()
	}
	if p.Branch != nil {
		m.Branch = &branchResourceModel{
			ID:           types.StringValue(p.Branch.ID),
			ProjectID:    types.StringValue(p.Branch.ProjectID),
			ParentID:     types.StringValue(p.Branch.ParentID),
			ParentLsn:    types.StringValue(p.Branch.ParentLsn),
			Name:         types.StringValue(p.Branch.Name),
			CurrentState: types.StringValue(p.Branch.CurrentState),
			CreatedAt:    types.StringValue(p.Branch.CreatedAt),
			UpdatedAt:    types.StringValue(p.Branch.UpdatedAt),
		}
	}
	if len(p.ConnectionUris) != 0 {
		m.ConnectionUris = []projectConnUris{}
		for _, v := range p.ConnectionUris {
			m.ConnectionUris = append(m.ConnectionUris, projectConnUris{
				ConnectionURI: types.StringValue(v.ConnectionURI),
			})
		}
	}
	if len(p.Roles) != 0 {
		m.Roles = []roleResourceModel{}
		for _, v := range p.Roles {
			m.Roles = append(m.Roles, roleResourceModel{
				BranchID:  types.StringValue(v.BranchID),
				Name:      types.StringValue(v.Name),
				Password:  types.StringValue(v.Password),
				Protected: types.BoolValue(v.Protected),
				CreatedAt: types.StringValue(v.CreatedAt),
				UpdatedAt: types.StringValue(v.UpdatedAt),
			})
		}
	}
	if len(p.Databases) != 0 {
		m.Databases = []databaseResourceModel{}
		for _, v := range p.Databases {
			m.Databases = append(m.Databases, databaseResourceModel{
				ID:        types.Int64Value(v.ID),
				BranchID:  types.StringValue(v.BranchID),
				Name:      types.StringValue(v.Name),
				OwnerName: types.StringValue(v.OwnerName),
				CreatedAt: types.StringValue(v.CreatedAt),
				UpdatedAt: types.StringValue(v.UpdatedAt),
			})
		}
	}
	if len(p.Endpoints) != 0 {
		m.Endpoints = []endpointResourceModel{}
		for _, v := range p.Endpoints {
			m.Endpoints = append(m.Endpoints, *v.ToEndpointResourceModel())
		}
	}
	return m
}
func (m *projectResourceModel) ToProjectResourceJSON() *projectResourceJSON {
	p := &projectResourceJSON{
		Project: innerProjectJSON{
			MaintenanceStartsAt:   m.Project.MaintenanceStartsAt.ValueString(),
			ID:                    m.Project.ID.ValueString(),
			PlatformID:            m.Project.PlatformID.ValueString(),
			RegionID:              m.Project.RegionID.ValueString(),
			Name:                  m.Project.Name.ValueString(),
			Provisioner:           m.Project.Provisioner.ValueString(),
			PgVersion:             m.Project.PgVersion.ValueInt64(),
			AutoscalingLimitMinCu: m.Project.AutoscalingLimitMinCu.ValueInt64(),
			AutoscalingLimitMaxCu: m.Project.AutoscalingLimitMaxCu.ValueInt64(),
			LastActive:            m.Project.LastActive.ValueString(),
			CreatedAt:             m.Project.CreatedAt.ValueString(),
			UpdatedAt:             m.Project.UpdatedAt.ValueString(),
		},
	}
	if m.Project.Settings != nil {
		p.Project.Settings = m.Project.Settings.ToEndpointSettingsJSON()
	}
	if m.Branch != nil {
		p.Branch = &branchResourceJSON{
			ID:           m.Branch.ID.ValueString(),
			ProjectID:    m.Branch.ProjectID.ValueString(),
			ParentID:     m.Branch.ParentID.ValueString(),
			ParentLsn:    m.Branch.ParentLsn.ValueString(),
			Name:         m.Branch.Name.ValueString(),
			CurrentState: m.Branch.CurrentState.ValueString(),
			CreatedAt:    m.Branch.CreatedAt.ValueString(),
			UpdatedAt:    m.Branch.UpdatedAt.ValueString(),
		}
	}
	if len(m.ConnectionUris) != 0 {
		p.ConnectionUris = []projectConnUrisJSON{}
		for _, v := range m.ConnectionUris {
			p.ConnectionUris = append(p.ConnectionUris, projectConnUrisJSON{
				ConnectionURI: v.ConnectionURI.ValueString(),
			})
		}
	}
	if len(m.Roles) != 0 {
		p.Roles = []roleResourceJSON{}
		for _, v := range m.Roles {
			p.Roles = append(p.Roles, roleResourceJSON{
				BranchID:  v.BranchID.ValueString(),
				Name:      v.Name.ValueString(),
				Password:  v.Password.ValueString(),
				Protected: v.Protected.ValueBool(),
				CreatedAt: v.UpdatedAt.ValueString(),
				UpdatedAt: v.UpdatedAt.ValueString(),
			})
		}
	}
	if len(m.Databases) != 0 {
		p.Databases = []databaseResourceJSON{}
		for _, v := range m.Databases {
			p.Databases = append(p.Databases, databaseResourceJSON{
				ID:        v.ID.ValueInt64(),
				BranchID:  v.BranchID.ValueString(),
				Name:      v.Name.ValueString(),
				OwnerName: v.OwnerName.ValueString(),
				CreatedAt: v.CreatedAt.ValueString(),
				UpdatedAt: v.UpdatedAt.ValueString(),
			})
		}
	}
	if len(m.Endpoints) != 0 {
		p.Endpoints = []endpointResourceJSON{}
		for _, v := range m.Endpoints {
			p.Endpoints = append(p.Endpoints, *v.ToEndpointResourceJSON())
		}
	}
	return p
}
