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

package wif_config

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/redhat/terraform-provider-osd-google/provider/common"
)

// WifConfigResource implements the osdgoogle_wif_config resource.
type WifConfigResource struct {
	wifConfigs *cmv1.WifConfigsClient
}

var _ resource.Resource = &WifConfigResource{}
var _ resource.ResourceWithConfigure = &WifConfigResource{}
var _ resource.ResourceWithImportState = &WifConfigResource{}

// New creates a new WIF config resource.
func New() resource.Resource {
	return &WifConfigResource{}
}

func (r *WifConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_wif_config"
}

func (r *WifConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Workload Identity Federation (WIF) configuration for OSD clusters on GCP.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the WIF config.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Description: "Human-readable display name for the WIF config.",
				Required:    true,
			},
			"organization": schema.StringAttribute{
				Description: "OCM organization ID owning this config.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"gcp": schema.SingleNestedAttribute{
				Description: "GCP-specific WIF configuration.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"project_id": schema.StringAttribute{
						Description: "GCP project ID where WIF resources are configured.",
						Required:    true,
					},
					"project_number": schema.StringAttribute{
						Description: "GCP project number for WIF resources.",
						Required:    true,
					},
					"role_prefix": schema.StringAttribute{
						Description: "Prefix for GCP custom role names.",
						Required:    true,
					},
					"federated_project_id": schema.StringAttribute{
						Description: "GCP project ID where WorkloadIdentityPool resources are configured (if different).",
						Optional:    true,
					},
					"federated_project_number": schema.StringAttribute{
						Description: "GCP project number for WorkloadIdentityPool.",
						Optional:    true,
					},
					"impersonator_email": schema.StringAttribute{
						Description: "Service account email used by OCM to access other service accounts.",
						Optional:    true,
					},
				},
			},
		},
	}
}

func (r *WifConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	conn, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError("unexpected provider data type", fmt.Sprintf("expected *sdk.Connection, got %T", req.ProviderData))
		return
	}
	r.wifConfigs = conn.ClustersMgmt().V1().GCP().WifConfigs()
}

func (r *WifConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WifConfigState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wifObj, err := r.buildWifConfigObject(&plan)
	if err != nil {
		resp.Diagnostics.AddError("failed to build WIF config", err.Error())
		return
	}

	addResp, err := r.wifConfigs.Add().Body(wifObj).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to create WIF config", err.Error())
		return
	}
	obj := addResp.Body()

	r.populateState(obj, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WifConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WifConfigState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	getResp, err := r.wifConfigs.WifConfig(state.ID.ValueString()).Get().SendContext(ctx)
	if err != nil {
		if getResp != nil && getResp.Status() == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("failed to get WIF config", err.Error())
		return
	}
	obj := getResp.Body()

	r.populateState(obj, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WifConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan WifConfigState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	wifObj, err := r.buildWifConfigObject(&plan)
	if err != nil {
		resp.Diagnostics.AddError("failed to build WIF config", err.Error())
		return
	}

	_, err = r.wifConfigs.WifConfig(plan.ID.ValueString()).Update().Body(wifObj).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to update WIF config", err.Error())
		return
	}

	getResp, err := r.wifConfigs.WifConfig(plan.ID.ValueString()).Get().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to get WIF config after update", err.Error())
		return
	}
	r.populateState(getResp.Body(), &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WifConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WifConfigState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.wifConfigs.WifConfig(state.ID.ValueString()).Delete().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete WIF config", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *WifConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *WifConfigResource) buildWifConfigObject(s *WifConfigState) (*cmv1.WifConfig, error) {
	gcpBuilder := cmv1.NewWifGcp().
		ProjectId(s.GCP.ProjectID.ValueString()).
		ProjectNumber(s.GCP.ProjectNumber.ValueString()).
		RolePrefix(s.GCP.RolePrefix.ValueString())

	if common.HasValue(s.GCP.FederatedProjectID) {
		gcpBuilder.FederatedProjectId(s.GCP.FederatedProjectID.ValueString())
	}
	if common.HasValue(s.GCP.FederatedProjectNumber) {
		gcpBuilder.FederatedProjectNumber(s.GCP.FederatedProjectNumber.ValueString())
	}
	if common.HasValue(s.GCP.ImpersonatorEmail) {
		gcpBuilder.ImpersonatorEmail(s.GCP.ImpersonatorEmail.ValueString())
	}

	return cmv1.NewWifConfig().
		DisplayName(s.DisplayName.ValueString()).
		Gcp(gcpBuilder).
		Build()
}

func (r *WifConfigResource) populateState(obj *cmv1.WifConfig, state *WifConfigState) {
	state.ID = types.StringValue(obj.ID())
	state.DisplayName = types.StringValue(obj.DisplayName())
	if obj.Organization() != nil {
		state.Organization = types.StringValue(obj.Organization().ID())
	}
	if obj.Gcp() != nil {
		gcp := obj.Gcp()
		state.GCP = &WifGcpState{
			ProjectID:     types.StringValue(gcp.ProjectId()),
			ProjectNumber: types.StringValue(gcp.ProjectNumber()),
			RolePrefix:    types.StringValue(gcp.RolePrefix()),
		}
		if gcp.FederatedProjectId() != "" {
			state.GCP.FederatedProjectID = types.StringValue(gcp.FederatedProjectId())
		}
		if gcp.FederatedProjectNumber() != "" {
			state.GCP.FederatedProjectNumber = types.StringValue(gcp.FederatedProjectNumber())
		}
		if gcp.ImpersonatorEmail() != "" {
			state.GCP.ImpersonatorEmail = types.StringValue(gcp.ImpersonatorEmail())
		}
	}
}
