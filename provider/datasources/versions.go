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

// VersionsDataSource implements osdgoogle_versions data source.
type VersionsDataSource struct {
	collection *cmv1.VersionsClient
}

var _ datasource.DataSource = &VersionsDataSource{}
var _ datasource.DataSourceWithConfigure = &VersionsDataSource{}

// NewVersions creates a new versions data source.
func NewVersions() datasource.DataSource {
	return &VersionsDataSource{}
}

func (d *VersionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_versions"
}

func (d *VersionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "List of available OpenShift versions for OSD.",
		Attributes: map[string]schema.Attribute{
			"items": schema.ListNestedAttribute{
				Description: "List of versions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":   schema.StringAttribute{Computed: true, Description: "Version ID (e.g., openshift-v4.16.1)."},
						"name": schema.StringAttribute{Computed: true, Description: "Short name (e.g., 4.16.1)."},
					},
				},
			},
		},
	}
}

func (d *VersionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	conn, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
	d.collection = conn.ClustersMgmt().V1().Versions()
}

func (d *VersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var items []VersionItem
	listReq := d.collection.List().Size(100).Search("enabled = 't'")
	for page := 1; ; page++ {
		listReq.Page(page)
		listResp, err := listReq.SendContext(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to list versions", err.Error())
			return
		}
		listResp.Items().Each(func(v *cmv1.Version) bool {
			name := v.ID()
			if rawID, ok := v.GetRawID(); ok {
				name = rawID
			}
			items = append(items, VersionItem{
				ID:   types.StringValue(v.ID()),
				Name: types.StringValue(name),
			})
			return true
		})
		if listResp.Size() < 100 {
			break
		}
	}
	state := VersionsState{Items: items}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

type VersionsState struct {
	Items []VersionItem `tfsdk:"items"`
}

type VersionItem struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
