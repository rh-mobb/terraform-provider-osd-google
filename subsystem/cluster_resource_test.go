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
	. "github.com/redhat/terraform-provider-osd-google/subsystem/framework"
)

var _ = Describe("Cluster resource", func() {
	// Pending: requires OCM mock server with full cluster API response format.
	// The provider Create flow needs properly shaped JSON for populateState.
	PIt("Can create and destroy a cluster", func() {
		// Mock cluster create response
		createResp := `{
		  "id": "cluster-123",
		  "name": "test-cluster",
		  "state": "installing",
		  "product": {"id": "osd"},
		  "cloud_provider": {"id": "gcp"},
		  "region": {"id": "us-central1"},
		  "nodes": {"compute": 3},
		  "version": {"id": "openshift-v4.16.1"}
		}`

		getResp := `{
		  "id": "cluster-123",
		  "name": "test-cluster",
		  "state": "ready",
		  "product": {"id": "osd"},
		  "cloud_provider": {"id": "gcp"},
		  "region": {"id": "us-central1"},
		  "nodes": {"compute": 3},
		  "version": {"id": "openshift-v4.16.1"},
		  "api": {"url": "https://api.test.example.com"},
		  "console": {"url": "https://console.test.example.com"},
		  "domain": {"id": "test.example.com"},
		  "infrastructure": {"id": "test-infra"},
		  "ccs": {"enabled": true}
		}`

		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
				RespondWithJSON(http.StatusCreated, createResp),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/cluster-123"),
				RespondWithJSON(http.StatusOK, getResp),
			),
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/cluster-123"),
				RespondWithJSON(http.StatusNoContent, ""),
			),
		)

		Terraform.Source(`
		  resource "osdgoogle_cluster" "test" {
		    name             = "test-cluster"
		    cloud_region     = "us-central1"
		    gcp_project_id   = "my-gcp-project"
		    version          = "4.16.1"
		    compute_nodes   = 3
		    compute_machine_type = "custom-4-16384"
		    ccs_enabled     = true
		  }
		`)
		applyOutput := Terraform.Apply()
		Expect(applyOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("osdgoogle_cluster", "test")
		Expect(resource).To(MatchJQ(`.attributes.id`, "cluster-123"))
		Expect(resource).To(MatchJQ(`.attributes.name`, "test-cluster"))
		Expect(resource).To(MatchJQ(`.attributes.state`, "ready"))

		destroyOutput := Terraform.Destroy()
		Expect(destroyOutput.ExitCode).To(BeZero())
	})
})
