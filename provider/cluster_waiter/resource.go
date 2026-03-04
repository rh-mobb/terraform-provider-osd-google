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

package cluster_waiter

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/redhat/terraform-provider-osd-google/provider/common"
)

// ClusterWaiterResource implements the osdgoogle_cluster_waiter resource.
type ClusterWaiterResource struct {
	collection  *cmv1.ClustersClient
	clusterWait common.ClusterWait
}

var _ resource.Resource = &ClusterWaiterResource{}
var _ resource.ResourceWithConfigure = &ClusterWaiterResource{}

const defaultTimeoutMinutes = 60

// New creates a new cluster waiter resource.
func New() resource.Resource {
	return &ClusterWaiterResource{}
}

func (r *ClusterWaiterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_waiter"
}

func (r *ClusterWaiterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Waits for an OSD cluster to become ready. Does not manage any infrastructure.",
		Attributes: map[string]schema.Attribute{
			"cluster_id": schema.StringAttribute{
				Description: "ID of the cluster to wait for.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"timeout": schema.Int64Attribute{
				Description: "Timeout in minutes for the wait. Default is 60.",
				Optional:    true,
			},
			"ready": schema.BoolAttribute{
				Description: "True when the cluster is ready.",
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ClusterWaiterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	conn, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError("unexpected provider data type", fmt.Sprintf("expected *sdk.Connection, got %T", req.ProviderData))
		return
	}
	r.collection = conn.ClustersMgmt().V1().Clusters()
	r.clusterWait = common.NewClusterWait(r.collection, conn)
}

func (r *ClusterWaiterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var state ClusterWaiterState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	object, err := r.startPolling(ctx, &state)
	if err != nil {
		resp.Diagnostics.AddError("cluster wait failed", err.Error())
		if object != nil {
			resp.Diagnostics.Append(resp.State.Set(ctx, object)...)
		}
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, object)...)
}

func (r *ClusterWaiterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// No-op - the cluster waiter doesn't store server state
}

func (r *ClusterWaiterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ClusterWaiterState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	object, err := r.startPolling(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("cluster wait failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, object)...)
}

func (r *ClusterWaiterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *ClusterWaiterResource) startPolling(ctx context.Context, state *ClusterWaiterState) (*ClusterWaiterState, error) {
	state.Ready = types.BoolValue(false)
	timeout := int64(defaultTimeoutMinutes)
	if !state.Timeout.IsNull() && !state.Timeout.IsUnknown() && state.Timeout.ValueInt64() > 0 {
		timeout = state.Timeout.ValueInt64()
	}
	object, err := r.clusterWait.WaitForClusterToBeReady(ctx, state.ClusterID.ValueString(), timeout)
	if err != nil {
		return state, fmt.Errorf("waiting for cluster %s: %w", state.ClusterID.ValueString(), err)
	}
	state.Ready = types.BoolValue(object.State() == cmv1.ClusterStateReady)
	return state, nil
}
