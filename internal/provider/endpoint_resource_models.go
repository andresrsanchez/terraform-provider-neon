package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type endpointResourceJSON struct {
	Host                  string                `json:"host"`
	Id                    string                `json:"id"`
	ProjectID             string                `json:"project_id"`
	BranchID              string                `json:"branch_id"`
	AutoscalingLimitMinCu int64                 `json:"autoscaling_limit_min_cu"`
	AutoscalingLimitMaxCu int64                 `json:"autoscaling_limit_max_cu"`
	RegionID              string                `json:"region_id"`
	Type                  string                `json:"type"`
	CurrentState          string                `json:"current_state"`
	PendingState          string                `json:"pending_state"`
	Settings              *endpointSettingsJSON `json:"settings"`
	PoolerEnabled         bool                  `json:"pooler_enabled"`
	PoolerMode            string                `json:"pooler_mode"`
	Disabled              bool                  `json:"disabled"`
	PasswordlessAccess    bool                  `json:"passwordless_access"`
	LastActive            string                `json:"last_active"`
	CreatedAt             string                `json:"created_at"`
	UpdatedAt             string                `json:"updated_at"`
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
	//Settings              types.Object `tfsdk:"settings"`
	PoolerEnabled      types.Bool   `tfsdk:"pooler_enabled"`
	PoolerMode         types.String `tfsdk:"pooler_mode"`
	Disabled           types.Bool   `tfsdk:"disabled"`
	PasswordlessAccess types.Bool   `tfsdk:"passwordless_access"`
	LastActive         types.String `tfsdk:"last_active"`
	CreatedAt          types.String `tfsdk:"created_at"`
	UpdatedAt          types.String `tfsdk:"updated_at"`
}

func (m *endpointResourceModel) ToEndpointResourceJSON() *endpointResourceJSON {
	e := &endpointResourceJSON{
		Host:                  m.Host.ValueString(),
		Id:                    m.Id.ValueString(),
		ProjectID:             m.ProjectID.ValueString(),
		BranchID:              m.BranchID.ValueString(),
		AutoscalingLimitMinCu: m.AutoscalingLimitMinCu.ValueInt64(),
		AutoscalingLimitMaxCu: m.AutoscalingLimitMaxCu.ValueInt64(),
		RegionID:              m.RegionID.ValueString(),
		Type:                  m.Type.ValueString(),
		CurrentState:          m.CurrentState.ValueString(),
		PendingState:          m.PendingState.ValueString(),
		PoolerEnabled:         m.PoolerEnabled.ValueBool(),
		PoolerMode:            m.PoolerMode.ValueString(),
		Disabled:              m.Disabled.ValueBool(),
		PasswordlessAccess:    m.PasswordlessAccess.ValueBool(),
		LastActive:            m.LastActive.ValueString(),
		CreatedAt:             m.CreatedAt.ValueString(),
		UpdatedAt:             m.UpdatedAt.ValueString(),
	}
	/*if !m.Settings.IsNull() {
		v := endpointSettingsModel{}
		m.Settings.As(context.TODO(), &v, types.ObjectAsOptions{})
		e.Settings = v.ToEndpointSettingsJSON()
	}*/
	return e
}

func (m *endpointResourceJSON) ToEndpointResourceModel() *endpointResourceModel {
	e := &endpointResourceModel{
		Host:                  types.StringValue(m.Host),
		Id:                    types.StringValue(m.Id),
		ProjectID:             types.StringValue(m.ProjectID),
		BranchID:              types.StringValue(m.BranchID),
		AutoscalingLimitMinCu: types.Int64Value(m.AutoscalingLimitMinCu),
		AutoscalingLimitMaxCu: types.Int64Value(m.AutoscalingLimitMaxCu),
		RegionID:              types.StringValue(m.RegionID),
		Type:                  types.StringValue(m.Type),
		CurrentState:          types.StringValue(m.CurrentState),
		PendingState:          types.StringValue(m.PendingState),
		PoolerEnabled:         types.BoolValue(m.PoolerEnabled),
		PoolerMode:            types.StringValue(m.PoolerMode),
		Disabled:              types.BoolValue(m.Disabled),
		PasswordlessAccess:    types.BoolValue(m.PasswordlessAccess),
		LastActive:            types.StringValue(m.LastActive),
		CreatedAt:             types.StringValue(m.CreatedAt),
		UpdatedAt:             types.StringValue(m.UpdatedAt),
	}
	/*if m.Settings != nil {
		pg, _ := types.MapValueFrom(context.TODO(), types.StringType, m.Settings.PgSettings)
		settings := map[string]attr.Value{
			"pg_settings": pg,
		}
		aux, _ := types.ObjectValueFrom(context.TODO(), typeFromAttrs(defaultSettingsResourceAttr()), settings)
		e.Settings = aux
	}*/
	return e
}

func (m *endpointSettingsJSON) ToEndpointSettingsModel() (*endpointSettingsModel, diag.Diagnostics) {
	settings, diags := types.MapValueFrom(context.TODO(), types.StringType, m.PgSettings)
	return &endpointSettingsModel{
		//Description: types.StringValue(m.Description),
		PgSettings: settings,
	}, diags
}

func (m *endpointSettingsModel) ToEndpointSettingsJSON() *endpointSettingsJSON {
	s := make(map[string]string)
	for k, v := range m.PgSettings.Elements() {
		pg_settings := v.(types.String)
		s[k] = pg_settings.ValueString()
	}
	return &endpointSettingsJSON{
		PgSettings: s,
	}
}
