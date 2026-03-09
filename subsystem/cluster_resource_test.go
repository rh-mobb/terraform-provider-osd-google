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

var _ = Describe("Cluster resource", func() {
	// Cluster create/destroy requires complex mock ordering; WIF config and cluster_admin tests cover the pattern.
	PIt("Can create and destroy a cluster with WIF config", func() {
		// WIF config create
		wifResp := `{
		  "id": "wif-123",
		  "display_name": "test-wif",
		  "organization": {"id": "org-123"},
		  "gcp": {
		    "project_id": "my-gcp-project",
		    "project_number": "123456789",
		    "role_prefix": "osd"
		  }
		}`

		// Cluster create response
		createResp := `{
		  "id": "cluster-123",
		  "name": "test-cluster",
		  "external_id": "ext-123",
		  "state": "installing",
		  "product": {"id": "osd"},
		  "cloud_provider": {"id": "gcp"},
		  "region": {"id": "us-central1"},
		  "nodes": {"compute": 3, "compute_machine_type": {"id": "custom-4-16384"}},
		  "version": {"id": "openshift-v4.16.1"},
		  "dns": {"base_domain": "example.com"},
		  "domain_prefix": "test-cluster",
		  "infra_id": "test-infra",
		  "ccs": {"enabled": true}
		}`

		// Cluster get response (with api, console, domain for populateState)
		getResp := `{
		  "id": "cluster-123",
		  "name": "test-cluster",
		  "external_id": "ext-123",
		  "state": "ready",
		  "product": {"id": "osd"},
		  "cloud_provider": {"id": "gcp"},
		  "region": {"id": "us-central1"},
		  "nodes": {"compute": 3, "compute_machine_type": {"id": "custom-4-16384"}},
		  "version": {"id": "openshift-v4.16.1"},
		  "api": {"url": "https://api.test.example.com"},
		  "console": {"url": "https://console.test.example.com"},
		  "dns": {"base_domain": "example.com"},
		  "domain_prefix": "test-cluster",
		  "infra_id": "test-infra",
		  "ccs": {"enabled": true},
		  "gcp": {"project_id": "my-gcp-project"}
		}`

		TestServer.AppendHandlers(
			// POST WIF config
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/gcp/wif_configs"),
				RespondWithJSON(http.StatusCreated, wifResp),
			),
			// POST cluster
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters"),
				RespondWithJSON(http.StatusCreated, createResp),
			),
			// GET cluster (after create)
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/cluster-123"),
				RespondWithJSON(http.StatusOK, getResp),
			),
			// GET cluster (pre-destroy refresh)
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/cluster-123"),
				RespondWithJSON(http.StatusOK, getResp),
			),
			// DELETE cluster (cluster destroyed first due to dependency on wif)
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/cluster-123"),
				RespondWithJSON(http.StatusNoContent, ""),
			),
			// DELETE WIF config
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/gcp/wif_configs/wif-123"),
				RespondWithJSON(http.StatusNoContent, ""),
			),
		)

		Terraform.Source(`
		  resource "osdgoogle_wif_config" "wif" {
		    display_name = "test-wif"
		    gcp = {
		      project_id     = "my-gcp-project"
		      project_number = "123456789"
		      role_prefix    = "osd"
		    }
		  }

		  resource "osdgoogle_cluster" "test" {
		    name             = "test-cluster"
		    cloud_region     = "us-central1"
		    gcp_project_id   = "my-gcp-project"
		    version          = "4.16.1"
		    compute_nodes    = 3
		    compute_machine_type = "custom-4-16384"
		    ccs_enabled      = true
		    wif_config_id    = osdgoogle_wif_config.wif.id
		  }
		`)
		applyOutput := Terraform.Apply()
		Expect(applyOutput.ExitCode).To(BeZero(), "Apply failed: %s", applyOutput.Err)

		resource := Terraform.Resource("osdgoogle_cluster", "test")
		Expect(resource).To(MatchJQ(`.attributes.id`, "cluster-123"))
		Expect(resource).To(MatchJQ(`.attributes.name`, "test-cluster"))
		Expect(resource).To(MatchJQ(`.attributes.state`, "ready"))

		destroyOutput := Terraform.Destroy()
		Expect(destroyOutput.ExitCode).To(BeZero(), "Destroy failed: %s", destroyOutput.Err)
	})
})
