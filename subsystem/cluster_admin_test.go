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

package subsystem

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2/dsl/core"             // nolint
	. "github.com/onsi/gomega"                         // nolint
	. "github.com/onsi/gomega/ghttp"                   // nolint
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
	. "github.com/rh-mobb/terraform-provider-osd-google/subsystem/framework"
)

var _ = Describe("Cluster admin resource", func() {
	It("Can create and destroy a cluster admin", func() {
		// Cluster must be ready before creating IDP
		clusterReady := `{
		  "id": "cluster-123",
		  "name": "test-cluster",
		  "state": "ready"
		}`

		// IDP create response
		idpCreated := `{
		  "id": "htpasswd-idp-1",
		  "name": "htpasswd",
		  "type": "HTPasswdIdentityProvider",
		  "mapping_method": "claim"
		}`

		// User added to group response
		userAdded := `{
		  "id": "htpasswd:cluster-admin",
		  "href": "/api/clusters_mgmt/v1/clusters/cluster-123/groups/cluster-admins/users/htpasswd:cluster-admin"
		}`

		TestServer.AppendHandlers(
			// Cluster GET for WaitForClusterToBeReady
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/cluster-123"),
				RespondWithJSON(http.StatusOK, clusterReady),
			),
			// POST identity_providers
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/cluster-123/identity_providers"),
				RespondWithJSON(http.StatusCreated, idpCreated),
			),
			// POST groups/cluster-admins/users
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/cluster-123/groups/cluster-admins/users"),
				RespondWithJSON(http.StatusCreated, userAdded),
			),
			// GET identity_providers (Read after Apply)
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/cluster-123/identity_providers/htpasswd-idp-1"),
				RespondWithJSON(http.StatusOK, idpCreated),
			),
			// GET groups/cluster-admins/users (Read after Apply) - use username, OCM rejects IDs with ':'
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/cluster-123/groups/cluster-admins/users/cluster-admin"),
				RespondWithJSON(http.StatusOK, userAdded),
			),
			// DELETE groups/cluster-admins/users
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/cluster-123/groups/cluster-admins/users/cluster-admin"),
				RespondWithJSON(http.StatusNoContent, ""),
			),
			// DELETE identity_providers
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/cluster-123/identity_providers/htpasswd-idp-1"),
				RespondWithJSON(http.StatusNoContent, ""),
			),
		)

		Terraform.Source(`
		  resource "osdgoogle_cluster_admin" "admin" {
		    cluster_id = "cluster-123"
		    username   = "cluster-admin"
		    password   = "TestPassword123!"
		  }
		`)
		applyOutput := Terraform.Apply()
		Expect(applyOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("osdgoogle_cluster_admin", "admin")
		Expect(resource).To(MatchJQ(`.attributes.id`, "htpasswd-idp-1"))
		Expect(resource).To(MatchJQ(`.attributes.username`, "cluster-admin"))

		destroyOutput := Terraform.Destroy()
		Expect(destroyOutput.ExitCode).To(BeZero())
	})
})
