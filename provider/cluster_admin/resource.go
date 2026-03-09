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

package cluster_admin

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk "github.com/openshift-online/ocm-sdk-go"
	cmv1 "github.com/openshift-online/ocm-sdk-go/clustersmgmt/v1"

	"github.com/rh-mobb/terraform-provider-osd-google/provider/common"
)

const (
	defaultClusterAdminUsername = "admin"
	clusterAdminsGroupID       = "cluster-admins"
	htpasswdIDPName            = "htpasswd"
	defaultWaitTimeoutMinutes  = 60
	passwordLength             = 23
)

// ClusterAdminResource implements the osdgoogle_cluster_admin resource.
type ClusterAdminResource struct {
	collection  *cmv1.ClustersClient
	clusterWait common.ClusterWait
}

var _ resource.Resource = &ClusterAdminResource{}
var _ resource.ResourceWithConfigure = &ClusterAdminResource{}
var _ resource.ResourceWithImportState = &ClusterAdminResource{}

// New creates a new cluster admin resource.
func New() resource.Resource {
	return &ClusterAdminResource{}
}

func (r *ClusterAdminResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_admin"
}

func (r *ClusterAdminResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Creates an HTPasswd identity provider with a cluster-admin user for an OSD cluster on GCP.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "OCM identity provider ID.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Description: "Identifier of the cluster.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Description: "Admin username. Defaults to 'admin'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(defaultClusterAdminUsername),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Description: "Admin password. Auto-generated if omitted. Must contain uppercase, lowercase, digit, and special character (min 14 chars).",
				Optional:    true,
				Sensitive:   true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ClusterAdminResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	conn, ok := req.ProviderData.(*sdk.Connection)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *sdk.Connection, got: %T. Please report this issue to the provider developers.", req.ProviderData))
		return
	}
	r.collection = conn.ClustersMgmt().V1().Clusters()
	r.clusterWait = common.NewClusterWait(r.collection, conn)
}

func (r *ClusterAdminResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ClusterAdminState
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := plan.ClusterID.ValueString()
	username := plan.Username.ValueString()
	if username == "" {
		username = defaultClusterAdminUsername
	}

	// Wait for cluster to be ready
	_, err := r.clusterWait.WaitForClusterToBeReady(ctx, clusterID, defaultWaitTimeoutMinutes)
	if err != nil {
		resp.Diagnostics.AddError("cluster not ready", fmt.Sprintf("cluster %s: %v", clusterID, err))
		return
	}

	// Determine password
	password := plan.Password.ValueString()
	if password == "" {
		var genErr error
		password, genErr = generatePassword()
		if genErr != nil {
			resp.Diagnostics.AddError("failed to generate password", genErr.Error())
			return
		}
	}

	// Create HTPasswd IDP with username and password
	htpasswdBuilder := cmv1.NewHTPasswdIdentityProvider().
		Username(username).
		Password(password)

	idpBuilder, err := cmv1.NewIdentityProvider().
		Name(htpasswdIDPName).
		Type(cmv1.IdentityProviderTypeHtpasswd).
		Htpasswd(htpasswdBuilder).
		MappingMethod(cmv1.IdentityProviderMappingMethodClaim).
		Login(true).
		Build()
	if err != nil {
		resp.Diagnostics.AddError("failed to build identity provider", err.Error())
		return
	}

	idpResp, err := r.collection.Cluster(clusterID).IdentityProviders().Add().Body(idpBuilder).SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to create identity provider", err.Error())
		return
	}
	idp := idpResp.Body()
	idpID := idp.ID()

	// Add user to cluster-admins group. OCM rejects user IDs containing ':',
	// so use the username only (not idp:username).
	userBuilder, err := cmv1.NewUser().ID(username).Build()
	if err != nil {
		resp.Diagnostics.AddError("failed to build user", err.Error())
		return
	}

	_, err = r.collection.Cluster(clusterID).
		Groups().Group(clusterAdminsGroupID).
		Users().Add().Body(userBuilder).SendContext(ctx)
	if err != nil {
		// Best-effort rollback: delete the IDP we just created
		_, _ = r.collection.Cluster(clusterID).IdentityProviders().IdentityProvider(idpID).Delete().SendContext(ctx)
		resp.Diagnostics.AddError("failed to add user to cluster-admins group", err.Error())
		return
	}

	// Populate state
	plan.ID = types.StringValue(idpID)
	plan.ClusterID = plan.ClusterID
	plan.Username = types.StringValue(username)
	plan.Password = types.StringValue(password)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ClusterAdminResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ClusterAdminState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.ClusterID.ValueString()
	idpID := state.ID.ValueString()
	username := state.Username.ValueString()
	if username == "" {
		username = defaultClusterAdminUsername
	}

	// Check IDP exists
	idpResp, err := r.collection.Cluster(clusterID).IdentityProviders().IdentityProvider(idpID).Get().SendContext(ctx)
	if err != nil {
		if idpResp != nil && idpResp.Status() == http.StatusNotFound {
			common.HandleNotFound(ctx, resp, "cluster_admin", clusterID+"/"+idpID)
			return
		}
		resp.Diagnostics.AddError("failed to get identity provider", err.Error())
		return
	}

	// Check user is in cluster-admins group (use username; OCM rejects IDs with ':')
	userResp, err := r.collection.Cluster(clusterID).
		Groups().Group(clusterAdminsGroupID).
		Users().User(username).Get().SendContext(ctx)
	if err != nil {
		if userResp != nil && userResp.Status() == http.StatusNotFound {
			common.HandleNotFound(ctx, resp, "cluster_admin", clusterID+"/"+username)
			return
		}
		resp.Diagnostics.AddError("failed to get group membership", err.Error())
		return
	}

	_ = idpResp.Body()
	_ = userResp.Body()

	// Password is not returned by API; keep existing state
	state.ID = types.StringValue(idpID)
	state.ClusterID = types.StringValue(clusterID)
	state.Username = types.StringValue(username)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ClusterAdminResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No in-place updates; all attributes are ForceNew or UseStateForUnknown
	resp.Diagnostics.Append(resp.State.Set(ctx, req.Plan)...)
}

func (r *ClusterAdminResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ClusterAdminState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := state.ClusterID.ValueString()
	idpID := state.ID.ValueString()
	username := state.Username.ValueString()
	if username == "" {
		username = defaultClusterAdminUsername
	}

	// Remove user from cluster-admins group (best-effort, ignore 404; use username)
	_, _ = r.collection.Cluster(clusterID).
		Groups().Group(clusterAdminsGroupID).
		Users().User(username).Delete().SendContext(ctx)

	// Delete identity provider
	_, err := r.collection.Cluster(clusterID).IdentityProviders().IdentityProvider(idpID).Delete().SendContext(ctx)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete identity provider", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *ClusterAdminResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Format: cluster_id/idp_id (idp_id is the identity provider ID from OCM)
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			"Import ID must be in the format 'cluster_id/idp_id'",
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// generatePassword creates a random password meeting OCM requirements:
// uppercase, lowercase, digit, special character, >= 14 chars (we use 23).
func generatePassword() (string, error) {
	const (
		upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lower   = "abcdefghijklmnopqrstuvwxyz"
		digits  = "0123456789"
		special = "!@#$%^&*-_=+"
		all     = upper + lower + digits + special
	)

	// Ensure at least one of each required character type
	buf := make([]byte, passwordLength)
	chars := []byte{upper[0], lower[0], digits[0], special[0]}
	for i := 0; i < 4; i++ {
		buf[i] = chars[i]
	}

	// Fill rest randomly
	allBytes := []byte(all)
	for i := 4; i < passwordLength; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(all))))
		if err != nil {
			return "", fmt.Errorf("crypto/rand: %w", err)
		}
		buf[i] = allBytes[n.Int64()]
	}

	// Shuffle to avoid predictable positions
	for i := len(buf) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", fmt.Errorf("crypto/rand: %w", err)
		}
		j := n.Int64()
		buf[i], buf[j] = buf[j], buf[i]
	}

	return string(buf), nil
}
