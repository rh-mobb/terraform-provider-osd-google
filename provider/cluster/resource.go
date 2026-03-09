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

package cluster

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/rh-mobb/terraform-provider-osd-google/provider/common"
)

const (
	versionPrefix  = "openshift-v"
	defaultProduct = "osd"
)

// ClusterResource implements the osdgoogle_cluster resource.
type ClusterResource struct {
	collection  *cmv1.ClustersClient
	clusterWait common.ClusterWait
	connection  *sdk.Connection
}

var _ resource.Resource = &ClusterResource{}
var _ resource.ResourceWithConfigure = &ClusterResource{}
var _ resource.ResourceWithImportState = &ClusterResource{}

// New creates a new cluster resource.
func New() resource.Resource {
	return &ClusterResource{}
}

func (r *ClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *ClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "OpenShift Dedicated (OSD) cluster on Google Cloud Platform. " +
			"CCS clusters require either wif_config_id (Workload Identity Federation) or gcp_authentication (service account key).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the cluster (from OCM).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"external_id": schema.StringAttribute{
				Description: "External identifier of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the cluster.",
				Required:    true,
			},
			"cloud_region": schema.StringAttribute{
				Description: "GCP region (e.g., us-central1).",
				Required:    true,
			},
			"gcp_project_id": schema.StringAttribute{
				Description: "GCP project ID for the cluster.",
				Required:    true,
			},
			"product": schema.StringAttribute{
				Description: "Product type (default: osd).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"multi_az": schema.BoolAttribute{
				Description: "Deploy across multiple availability zones.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"version": schema.StringAttribute{
				Description: "OpenShift version (e.g., 4.16.1).",
				Optional:    true,
			},
			"domain_prefix": schema.StringAttribute{
				Description: "DNS domain prefix for the cluster.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ccs_enabled": schema.BoolAttribute{
				Description: "Enable Customer Cloud Subscription (CCS) mode.",
				Optional:    true,
			},
			"billing_model": schema.StringAttribute{
				Description: "Billing model for the cluster. For CCS clusters, defaults to 'marketplace-gcp'. Set to 'standard' to use standard billing. Only 'standard' and 'marketplace-gcp' are allowed.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("standard", "marketplace-gcp"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"marketplace_gcp_terms": schema.BoolAttribute{
				Description: "Whether GCP marketplace terms have been accepted.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"compute_machine_type": schema.StringAttribute{
				Description: "GCP machine type for worker nodes (e.g., custom-4-16384). Defaults to n2-standard-4 when not specified.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"compute_nodes": schema.Int64Attribute{
				Description: "Number of worker nodes (when not using autoscaling).",
				Optional:    true,
			},
			"availability_zones": schema.ListAttribute{
				Description: "GCP availability zones for the cluster. Defaults to a single zone when not specified.",
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"properties": schema.MapAttribute{
				Description: "Cluster properties.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"wif_config_id": schema.StringAttribute{
				Description: "ID of the WIF config (when using Workload Identity Federation). Best practice: use one WIF config per cluster.",
				Optional:    true,
			},
			"wif_verify_timeout_minutes": schema.Int64Attribute{
				Description: "When using wif_config_id, wait up to this many minutes for WIF config verification before creating the cluster. GCP IAM propagation can take several minutes. Default 10.",
				Optional:    true,
			},
			"wait_for_create_complete": schema.BoolAttribute{
				Description: "Wait for cluster to be ready after creation. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"wait_timeout": schema.Int64Attribute{
				Description: "Timeout in minutes for cluster create and delete wait. Defaults to 60.",
				Optional:    true,
			},
			"gcp_authentication": schema.SingleNestedAttribute{
				Description: "GCP service account authentication (when not using WIF).",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"client_email":                schema.StringAttribute{Required: true, Sensitive: true},
					"client_id":                   schema.StringAttribute{Required: true},
					"private_key":                 schema.StringAttribute{Required: true, Sensitive: true},
					"private_key_id":              schema.StringAttribute{Required: true},
					"auth_uri":                    schema.StringAttribute{Optional: true},
					"token_uri":                   schema.StringAttribute{Optional: true},
					"auth_provider_x509_cert_url": schema.StringAttribute{Optional: true},
					"client_x509_cert_url":        schema.StringAttribute{Optional: true},
					"type":                        schema.StringAttribute{Optional: true},
				},
			},
			"private_service_connect": schema.SingleNestedAttribute{
				Description: "Private Service Connect configuration.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"service_attachment_subnet": schema.StringAttribute{Required: true},
				},
			},
			"gcp_network": schema.SingleNestedAttribute{
				Description: "GCP network (Shared VPC) configuration.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"vpc_name":             schema.StringAttribute{Required: true},
					"vpc_project_id":       schema.StringAttribute{Required: true},
					"compute_subnet":       schema.StringAttribute{Required: true},
					"control_plane_subnet": schema.StringAttribute{Required: true},
				},
			},
			"gcp_encryption_key": schema.SingleNestedAttribute{
				Description: "Customer-managed encryption key (CMEK) for CCS clusters.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"kms_key_service_account": schema.StringAttribute{Required: true},
					"key_location":            schema.StringAttribute{Required: true},
					"key_name":                schema.StringAttribute{Required: true},
					"key_ring":                schema.StringAttribute{Required: true},
				},
			},
			"security": schema.SingleNestedAttribute{
				Description: "GCP security settings.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"secure_boot": schema.BoolAttribute{Optional: true},
				},
			},
			"network": schema.SingleNestedAttribute{
				Description: "Network CIDR configuration.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"machine_cidr": schema.StringAttribute{Optional: true},
					"service_cidr": schema.StringAttribute{Optional: true},
					"pod_cidr":     schema.StringAttribute{Optional: true},
					"host_prefix":  schema.Int64Attribute{Optional: true},
				},
			},
			"autoscaling": schema.SingleNestedAttribute{
				Description: "Autoscaling configuration for worker nodes.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"min_replicas": schema.Int64Attribute{Required: true},
					"max_replicas": schema.Int64Attribute{Required: true},
				},
			},
			"proxy": schema.SingleNestedAttribute{
				Description: "Proxy configuration.",
				Optional:    true,
				Attributes: map[string]schema.Attribute{
					"http_proxy":              schema.StringAttribute{Optional: true},
					"https_proxy":             schema.StringAttribute{Optional: true},
					"no_proxy":                schema.StringAttribute{Optional: true},
					"additional_trust_bundle": schema.StringAttribute{Optional: true},
				},
			},
			"state": schema.StringAttribute{
				Description: "Current state of the cluster.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"api_url": schema.StringAttribute{
				Description: "API server URL.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"console_url": schema.StringAttribute{
				Description: "Web console URL.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain": schema.StringAttribute{
				Description: "Cluster domain.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"infra_id": schema.StringAttribute{
				Description: "Infrastructure ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"current_compute": schema.Int64Attribute{
				Description: "Current number of compute nodes.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	conn, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
	r.connection = conn
	r.collection = conn.ClustersMgmt().V1().Clusters()
	r.clusterWait = common.NewClusterWait(r.collection, conn)
}

func (r *ClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ClusterState
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.CCSEnabled.ValueBool() {
		hasWIF := plan.WIFConfigID.ValueString() != ""
		hasGCPAuth := plan.GCPAuthentication != nil
		if !hasWIF && !hasGCPAuth {
			resp.Diagnostics.AddError(
				"CCS cluster requires GCP credentials",
				"When ccs_enabled is true, you must provide either wif_config_id (Workload Identity Federation) or gcp_authentication (service account key). See the cluster docs for examples.",
			)
			return
		}
	}

	// When using WIF, wait for OCM to verify the GCP resources before cluster creation.
	// GCP IAM is eventually consistent; cluster creation fails with 503 until verification succeeds.
	if wifID := plan.WIFConfigID.ValueString(); wifID != "" {
		timeoutMin := plan.WifVerifyTimeoutMinutes.ValueInt64()
		if timeoutMin <= 0 {
			timeoutMin = 10
		}
		tflog.Info(ctx, fmt.Sprintf("Waiting for WIF config %s to be verified (timeout %d min)", wifID, timeoutMin))
		statusClient := r.connection.ClustersMgmt().V1().GCP().WifConfigs().WifConfig(wifID).Status()
		pollCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMin)*time.Minute)
		defer cancel()
		_, err := statusClient.Poll().
			Interval(30 * time.Second).
			Predicate(func(resp *cmv1.WifConfigStatusGetResponse) bool {
				body := resp.Body()
				return body != nil && body.Configured()
			}).
			StartContext(pollCtx)
		if err != nil {
			resp.Diagnostics.AddError(
				"WIF config not ready",
				fmt.Sprintf("Timed out waiting for WIF config %s to be verified. GCP IAM propagation can take several minutes. Run 'ocm gcp verify wif-config %s' to check status, or increase wif_verify_timeout_minutes.", wifID, wifID),
			)
			return
		}
		tflog.Info(ctx, "WIF config verified, proceeding with cluster creation")
	}

	clusterObj, err := r.buildClusterObject(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("failed to build cluster", err.Error())
		return
	}

	addResp, err := r.collection.Add().Body(clusterObj).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to create cluster", err.Error())
		return
	}
	cluster := addResp.Body()
	clusterID := cluster.ID()

	if plan.WaitForCreateComplete.ValueBool() {
		timeout := int64(60)
		if plan.WaitTimeout.ValueInt64() > 0 {
			timeout = plan.WaitTimeout.ValueInt64()
		}
		cluster, err = r.clusterWait.WaitForClusterToBeReady(ctx, clusterID, timeout)
		if err != nil {
			resp.Diagnostics.AddError("cluster creation wait failed", err.Error())
			return
		}
	} else {
		getResp, err := r.collection.Cluster(clusterID).Get().SendContext(ctx)
		if err != nil {
			resp.Diagnostics.AddError("failed to get cluster after create", err.Error())
			return
		}
		cluster = getResp.Body()
	}

	if err := r.populateState(ctx, cluster, &plan); err != nil {
		resp.Diagnostics.AddError("failed to populate state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ClusterState
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	getResp, err := r.collection.Cluster(state.ID.ValueString()).Get().SendContext(ctx)
	if err != nil {
		if getResp != nil && getResp.Status() == http.StatusNotFound {
			common.HandleNotFound(ctx, resp, "cluster", state.ID.ValueString())
			return
		}
		resp.Diagnostics.AddError("failed to get cluster", err.Error())
		return
	}
	cluster := getResp.Body()

	if err := r.populateState(ctx, cluster, &state); err != nil {
		resp.Diagnostics.AddError("failed to populate state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ClusterState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	patch, err := r.buildClusterPatch(&state, &plan)
	if err != nil {
		resp.Diagnostics.AddError("failed to build cluster patch", err.Error())
		return
	}
	if patch == nil {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	_, err = r.collection.Cluster(state.ID.ValueString()).Update().Body(patch).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to update cluster", err.Error())
		return
	}

	getResp, err := r.collection.Cluster(state.ID.ValueString()).Get().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to get cluster after update", err.Error())
		return
	}
	if err := r.populateState(ctx, getResp.Body(), &plan); err != nil {
		resp.Diagnostics.AddError("failed to populate state", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ClusterState
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.collection.Cluster(state.ID.ValueString()).Delete().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete cluster", err.Error())
		return
	}

	timeout := int64(60)
	if state.WaitTimeout.ValueInt64() > 0 {
		timeout = state.WaitTimeout.ValueInt64()
	}
	if err := r.clusterWait.WaitForClusterToBeDeleted(ctx, state.ID.ValueString(), timeout); err != nil {
		resp.Diagnostics.AddError("cluster deletion wait failed", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *ClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *ClusterResource) buildClusterObject(ctx context.Context, s *ClusterState) (*cmv1.Cluster, error) {
	builder := cmv1.NewCluster().
		Name(s.Name.ValueString()).
		CloudProvider(cmv1.NewCloudProvider().ID("gcp")).
		Region(cmv1.NewCloudRegion().ID(s.CloudRegion.ValueString())).
		Product(cmv1.NewProduct().ID(defaultProduct)).
		MultiAZ(s.MultiAZ.ValueBool())

	if s.Version.ValueString() != "" {
		versionID := s.Version.ValueString()
		if versionID != "" && len(versionID) < 20 && (len(versionID) < 1 || versionID[0] != 'o') {
			versionID = versionPrefix + versionID
		}
		builder.Version(cmv1.NewVersion().ID(versionID))
	}

	if s.DomainPrefix.ValueString() != "" {
		builder.DomainPrefix(s.DomainPrefix.ValueString())
	}

	ccsEnabled := s.CCSEnabled.ValueBool()
	builder.CCS(cmv1.NewCCS().Enabled(ccsEnabled))

	if ccsEnabled {
		bm := cmv1.BillingModelMarketplaceGCP
		if common.HasValue(s.BillingModel) && s.BillingModel.ValueString() == "standard" {
			bm = cmv1.BillingModelStandard
		}
		builder.BillingModel(bm)
	}

	if s.GCPProjectID.ValueString() != "" || common.HasValue(s.WIFConfigID) {
		gcpBuilder := cmv1.NewGCP()

		if common.HasValue(s.WIFConfigID) {
			// WIF clusters must not include project_id in the GCP body.
			gcpBuilder.Authentication(
				cmv1.NewGcpAuthentication().
					Kind(cmv1.WifConfigKind).
					Id(s.WIFConfigID.ValueString()),
			)
		} else {
			if s.GCPProjectID.ValueString() != "" {
				gcpBuilder.ProjectID(s.GCPProjectID.ValueString())
			}
			if s.GCPAuthentication != nil {
				auth := s.GCPAuthentication
				gcpBuilder.ClientEmail(auth.ClientEmail.ValueString()).
					ClientID(auth.ClientID.ValueString()).
					PrivateKey(auth.PrivateKey.ValueString()).
					PrivateKeyID(auth.PrivateKeyID.ValueString())
				if common.HasValue(auth.AuthURI) {
					gcpBuilder.AuthURI(auth.AuthURI.ValueString())
				}
				if common.HasValue(auth.TokenURI) {
					gcpBuilder.TokenURI(auth.TokenURI.ValueString())
				}
				if common.HasValue(auth.AuthProviderX509CertURL) {
					gcpBuilder.AuthProviderX509CertURL(auth.AuthProviderX509CertURL.ValueString())
				}
				if common.HasValue(auth.ClientX509CertURL) {
					gcpBuilder.ClientX509CertURL(auth.ClientX509CertURL.ValueString())
				}
				if common.HasValue(auth.Type) {
					gcpBuilder.Type(auth.Type.ValueString())
				}
			}
		}

		if s.PrivateServiceConnect != nil {
			gcpBuilder.PrivateServiceConnect(
				cmv1.NewGcpPrivateServiceConnect().ServiceAttachmentSubnet(s.PrivateServiceConnect.ServiceAttachmentSubnet.ValueString()),
			)
		}
		if s.Security != nil && !s.Security.SecureBoot.IsNull() {
			gcpBuilder.Security(cmv1.NewGcpSecurity().SecureBoot(s.Security.SecureBoot.ValueBool()))
		}
		builder.GCP(gcpBuilder)
	}

	if s.GCPNetwork != nil {
		netBuilder := cmv1.NewGCPNetwork().
			VPCName(s.GCPNetwork.VPCName.ValueString()).
			VPCProjectID(s.GCPNetwork.VPCProjectID.ValueString()).
			ComputeSubnet(s.GCPNetwork.ComputeSubnet.ValueString()).
			ControlPlaneSubnet(s.GCPNetwork.ControlPlaneSubnet.ValueString())
		builder.GCPNetwork(netBuilder)
	}

	if s.GCPEncryptionKey != nil {
		keyBuilder := cmv1.NewGCPEncryptionKey().
			KMSKeyServiceAccount(s.GCPEncryptionKey.KmsKeyServiceAccount.ValueString()).
			KeyLocation(s.GCPEncryptionKey.KeyLocation.ValueString()).
			KeyName(s.GCPEncryptionKey.KeyName.ValueString()).
			KeyRing(s.GCPEncryptionKey.KeyRing.ValueString())
		builder.GCPEncryptionKey(keyBuilder)
	}

	if s.Network != nil {
		netBuilder := cmv1.NewNetwork()
		if common.HasValue(s.Network.MachineCIDR) {
			netBuilder.MachineCIDR(s.Network.MachineCIDR.ValueString())
		}
		if common.HasValue(s.Network.ServiceCIDR) {
			netBuilder.ServiceCIDR(s.Network.ServiceCIDR.ValueString())
		}
		if common.HasValue(s.Network.PodCIDR) {
			netBuilder.PodCIDR(s.Network.PodCIDR.ValueString())
		}
		if !s.Network.HostPrefix.IsNull() {
			netBuilder.HostPrefix(int(s.Network.HostPrefix.ValueInt64()))
		}
		if !netBuilder.Empty() {
			builder.Network(netBuilder)
		}
	}

	nodesBuilder := cmv1.NewClusterNodes()
	if s.Autoscaling != nil {
		autoscaling := cmv1.NewMachinePoolAutoscaling().
			MinReplicas(int(s.Autoscaling.MinReplicas.ValueInt64())).
			MaxReplicas(int(s.Autoscaling.MaxReplicas.ValueInt64()))
		nodesBuilder.AutoscaleCompute(autoscaling)
	} else {
		computeNodes := int64(3)
		if !s.ComputeNodes.IsNull() {
			computeNodes = s.ComputeNodes.ValueInt64()
		}
		nodesBuilder.Compute(int(computeNodes))
	}
	if common.HasValue(s.ComputeMachineType) {
		nodesBuilder.ComputeMachineType(cmv1.NewMachineType().ID(s.ComputeMachineType.ValueString()))
	}
	if !s.AvailabilityZones.IsNull() && !s.AvailabilityZones.IsUnknown() {
		azs := common.StringListToArray(s.AvailabilityZones)
		nodesBuilder.AvailabilityZones(azs...)
	}
	builder.Nodes(nodesBuilder)

	if !s.Properties.IsNull() && !s.Properties.IsUnknown() {
		props := make(map[string]string)
		for k, v := range s.Properties.Elements() {
			if str, ok := v.(types.String); ok {
				props[k] = str.ValueString()
			}
		}
		builder.Properties(props)
	}

	if s.Proxy != nil {
		proxyBuilder := cmv1.NewProxy()
		if common.HasValue(s.Proxy.HTTPProxy) {
			proxyBuilder.HTTPProxy(s.Proxy.HTTPProxy.ValueString())
		}
		if common.HasValue(s.Proxy.HTTPSProxy) {
			proxyBuilder.HTTPSProxy(s.Proxy.HTTPSProxy.ValueString())
		}
		if common.HasValue(s.Proxy.NoProxy) {
			proxyBuilder.NoProxy(s.Proxy.NoProxy.ValueString())
		}
		if common.HasValue(s.Proxy.AdditionalTrustBundle) {
			// Proxy builder might use AdditionalTrustBundle - check API
			builder.AdditionalTrustBundle(s.Proxy.AdditionalTrustBundle.ValueString())
		}
		if !proxyBuilder.Empty() {
			builder.Proxy(proxyBuilder)
		}
	}

	return builder.Build()
}

func (r *ClusterResource) buildClusterPatch(state, plan *ClusterState) (*cmv1.Cluster, error) {
	updated := false
	builder := cmv1.NewCluster()

	if value, ok := common.ShouldPatchString(state.DomainPrefix, plan.DomainPrefix); ok {
		builder.DomainPrefix(value)
		updated = true
	}
	if value, ok := common.ShouldPatchBool(state.MultiAZ, plan.MultiAZ); ok {
		builder.MultiAZ(value)
		updated = true
	}
	if !plan.Properties.IsNull() && !plan.Properties.IsUnknown() {
		if _, ok := common.ShouldPatchMap(state.Properties, plan.Properties); ok {
			props := make(map[string]string)
			for k, v := range plan.Properties.Elements() {
				if str, ok := v.(types.String); ok {
					props[k] = str.ValueString()
				}
			}
			builder.Properties(props)
			updated = true
		}
	}

	if !updated {
		return nil, nil
	}
	return builder.Build()
}

func (r *ClusterResource) populateState(ctx context.Context, cluster *cmv1.Cluster, state *ClusterState) error {
	state.ID = types.StringValue(cluster.ID())
	state.ExternalID = types.StringValue(cluster.ExternalID())
	state.Name = types.StringValue(cluster.Name())
	state.CloudRegion = types.StringValue(cluster.Region().ID())
	state.Product = types.StringValue(cluster.Product().ID())
	state.MultiAZ = types.BoolValue(cluster.MultiAZ())
	state.DomainPrefix = types.StringValue(cluster.DomainPrefix())
	state.State = types.StringValue(string(cluster.State()))
	if cluster.DNS() != nil {
		state.Domain = types.StringValue(fmt.Sprintf("%s.%s", cluster.DomainPrefix(), cluster.DNS().BaseDomain()))
	} else {
		state.Domain = types.StringValue("")
	}
	state.InfraID = types.StringValue(cluster.InfraID())

	if cluster.API() != nil {
		state.APIURL = types.StringValue(cluster.API().URL())
	} else {
		state.APIURL = types.StringValue("")
	}
	if cluster.Console() != nil {
		state.ConsoleURL = types.StringValue(cluster.Console().URL())
	} else {
		state.ConsoleURL = types.StringValue("")
	}
	if cluster.Status() != nil {
		state.CurrentCompute = types.Int64Value(int64(cluster.Status().CurrentCompute()))
	} else {
		state.CurrentCompute = types.Int64Value(0)
	}

	if value, ok := cluster.GetBillingModel(); ok && string(value) != "" {
		state.BillingModel = types.StringValue(string(value))
	} else {
		state.BillingModel = types.StringValue("")
	}
	state.MarketplaceGCPTerms = types.BoolValue(false)

	if cluster.Nodes() != nil {
		state.ComputeNodes = types.Int64Value(int64(cluster.Nodes().Compute()))
		if cluster.Nodes().ComputeMachineType() != nil {
			state.ComputeMachineType = types.StringValue(cluster.Nodes().ComputeMachineType().ID())
		}
		if cluster.Nodes().AvailabilityZones() != nil {
			azList, diags := types.ListValueFrom(ctx, types.StringType, cluster.Nodes().AvailabilityZones())
			_ = diags
			state.AvailabilityZones = azList
		}
	}

	if cluster.GCP() != nil {
		state.GCPProjectID = types.StringValue(cluster.GCP().ProjectID())
	}

	return nil
}
