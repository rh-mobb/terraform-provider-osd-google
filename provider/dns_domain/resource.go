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

package dns_domain

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

	"github.com/rh-mobb/terraform-provider-osd-google/provider/common"
)

// DNSDomainResource implements the osdgoogle_dns_domain resource.
type DNSDomainResource struct {
	collection *cmv1.DNSDomainsClient
}

var _ resource.Resource = &DNSDomainResource{}
var _ resource.ResourceWithConfigure = &DNSDomainResource{}
var _ resource.ResourceWithImportState = &DNSDomainResource{}

// New creates a new DNS domain resource.
func New() resource.Resource {
	return &DNSDomainResource{}
}

func (r *DNSDomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_domain"
}

func (r *DNSDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reserves a DNS domain for an OSD cluster.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the DNS domain.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_arch": schema.StringAttribute{
				Description: "Cluster architecture for the DNS domain.",
				Optional:    true,
			},
		},
	}
}

func (r *DNSDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	conn, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
	r.collection = conn.ClustersMgmt().V1().DNSDomains()
}

func (r *DNSDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DNSDomainState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	builder := cmv1.NewDNSDomain()
	if !plan.ClusterArch.IsNull() && !plan.ClusterArch.IsUnknown() {
		builder.ClusterArch(cmv1.ClusterArchitecture(plan.ClusterArch.ValueString()))
	}
	payload, err := builder.Build()
	if err != nil {
		resp.Diagnostics.AddError("failed to build DNS domain", err.Error())
		return
	}
	addResp, err := r.collection.Add().Body(payload).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to create DNS domain", err.Error())
		return
	}
	obj := addResp.Body()
	plan.ID = types.StringValue(obj.ID())
	if arch, ok := obj.GetClusterArch(); ok {
		plan.ClusterArch = types.StringValue(string(arch))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *DNSDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DNSDomainState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	getResp, err := r.collection.DNSDomain(state.ID.ValueString()).Get().SendContext(ctx)
	if err != nil {
		if getResp != nil && getResp.Status() == http.StatusNotFound {
			common.HandleNotFound(ctx, resp, "dns_domain", state.ID.ValueString())
			return
		}
		resp.Diagnostics.AddError("failed to get DNS domain", err.Error())
		return
	}
	obj := getResp.Body()
	state.ID = types.StringValue(obj.ID())
	if arch, ok := obj.GetClusterArch(); ok {
		state.ClusterArch = types.StringValue(string(arch))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *DNSDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// DNS domain has no updateable fields in the basic schema
	resp.Diagnostics.Append(resp.State.Set(ctx, req.Plan)...)
}

func (r *DNSDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DNSDomainState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.collection.DNSDomain(state.ID.ValueString()).Delete().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete DNS domain", err.Error())
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *DNSDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
