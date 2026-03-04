/*
Copyright (c) 2025 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// MachineTypesDataSource implements osdgoogle_machine_types data source.
type MachineTypesDataSource struct {
	connection *sdk.Connection
}

var _ datasource.DataSource = &MachineTypesDataSource{}
var _ datasource.DataSourceWithConfigure = &MachineTypesDataSource{}

// NewMachineTypes creates a new machine types data source.
func NewMachineTypes() datasource.DataSource {
	return &MachineTypesDataSource{}
}

func (d *MachineTypesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine_types"
}

func (d *MachineTypesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List of GCP machine types available for OSD clusters in a region.",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Description: "GCP region (e.g., us-central1).",
				Required:    true,
			},
			"gcp_project_id": schema.StringAttribute{
				Description: "GCP project ID.",
				Required:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "List of machine types.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":   schema.StringAttribute{Computed: true, Description: "Machine type ID."},
						"name": schema.StringAttribute{Computed: true, Description: "Machine type name."},
					},
				},
			},
		},
	}
}

func (d *MachineTypesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	conn, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError("unexpected provider data type", "expected *sdk.Connection")
		return
	}
	d.connection = conn
}

func (d *MachineTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		Region       types.String `tfsdk:"region"`
		GCPProjectID types.String `tfsdk:"gcp_project_id"`
	}
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := cmv1.NewCloudProviderData().
		GCP(cmv1.NewGCP().ProjectID(config.GCPProjectID.ValueString())).
		Region(cmv1.NewCloudRegion().ID(config.Region.ValueString())).
		Build()
	if err != nil {
		resp.Diagnostics.AddError("failed to build request", err.Error())
		return
	}

	searchResp, err := d.connection.ClustersMgmt().V1().GCPInquiries().MachineTypes().Search().Body(body).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to list machine types", err.Error())
		return
	}

	var items []MachineTypeItem
	searchResp.Items().Each(func(mt *cmv1.MachineType) bool {
		items = append(items, MachineTypeItem{
			ID:   types.StringValue(mt.ID()),
			Name: types.StringValue(mt.Name()),
		})
		return true
	})

	state := MachineTypesState{Items: items, Region: config.Region, GCPProjectID: config.GCPProjectID}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

type MachineTypesState struct {
	Region       types.String     `tfsdk:"region"`
	GCPProjectID types.String     `tfsdk:"gcp_project_id"`
	Items        []MachineTypeItem `tfsdk:"items"`
}

type MachineTypeItem struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
