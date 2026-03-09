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
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"
)

// RegionsDataSource implements osdgoogle_regions data source.
type RegionsDataSource struct {
	connection *sdk.Connection
}

var _ datasource.DataSource = &RegionsDataSource{}
var _ datasource.DataSourceWithConfigure = &RegionsDataSource{}

// NewRegions creates a new regions data source.
func NewRegions() datasource.DataSource {
	return &RegionsDataSource{}
}

func (d *RegionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

func (d *RegionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List of GCP regions available for OSD clusters.",
		Attributes: map[string]schema.Attribute{
			"gcp_project_id": schema.StringAttribute{
				Description: "GCP project ID.",
				Required:    true,
			},
			"items": schema.ListNestedAttribute{
				Description: "List of regions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{Computed: true, Description: "Region ID (e.g., us-central1)."},
					},
				},
			},
		},
	}
}

func (d *RegionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	conn, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
	d.connection = conn
}

func (d *RegionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config RegionsState
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := cmv1.NewCloudProviderData().GCP(cmv1.NewGCP().ProjectID(config.GCPProjectID.ValueString())).Build()
	if err != nil {
		resp.Diagnostics.AddError("failed to build request", err.Error())
		return
	}

	searchResp, err := d.connection.ClustersMgmt().V1().GCPInquiries().Regions().Search().Body(body).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to list regions", err.Error())
		return
	}

	var items []RegionItem
	if searchResp.Items() != nil {
		searchResp.Items().Each(func(cr *cmv1.CloudRegion) bool {
			items = append(items, RegionItem{ID: types.StringValue(cr.ID())})
			return true
		})
	}

	state := RegionsState{Items: items, GCPProjectID: config.GCPProjectID}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

type RegionsState struct {
	GCPProjectID types.String  `tfsdk:"gcp_project_id"`
	Items        []RegionItem  `tfsdk:"items"`
}

type RegionItem struct {
	ID types.String `tfsdk:"id"`
}
