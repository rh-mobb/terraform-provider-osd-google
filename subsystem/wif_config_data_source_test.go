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

var wifConfigFullJSON = `{
  "id": "wif-config-123",
  "display_name": "my-wif",
  "organization": {"id": "org-123"},
  "gcp": {
    "project_id": "my-gcp-project",
    "project_number": "123456789",
    "role_prefix": "mywif",
    "federated_project_id": "my-gcp-project",
    "federated_project_number": "123456789",
    "impersonator_email": "ocm-impersonator@my-gcp-project.iam.gserviceaccount.com",
    "workload_identity_pool": {
      "pool_id": "mywif-pool",
      "identity_provider": {
        "identity_provider_id": "oidc",
        "issuer_url": "https://example.com/oidc",
        "jwks": "{}",
        "allowed_audiences": ["openshift"]
      }
    },
    "service_accounts": [
      {
        "service_account_id": "mywif-installer",
        "access_method": "impersonate",
        "osd_role": "installer",
        "roles": [{"role_id": "roles/compute.viewer", "predefined": true}]
      }
    ],
    "support": {
      "principal": "sd-sre-platform-gcp-access@redhat.com",
      "roles": [{"role_id": "roles/viewer", "predefined": true}]
    }
  }
}`

var _ = Describe("WIF config data source", func() {
	It("Can look up WIF config by display_name", func() {
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/gcp/wif_configs"),
				VerifyFormKV("search", "display_name = 'my-wif'"),
				VerifyFormKV("size", "2"),
				RespondWithJSON(http.StatusOK, `{
				  "items": [`+wifConfigFullJSON+`]
				}`),
			),
		)

		Terraform.Source(`
		  data "osdgoogle_wif_config" "wif" {
		    display_name = "my-wif"
		  }
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("osdgoogle_wif_config", "wif")
		Expect(resource).To(MatchJQ(`.attributes.id`, "wif-config-123"))
		Expect(resource).To(MatchJQ(`.attributes.display_name`, "my-wif"))
		Expect(resource).To(MatchJQ(`.attributes.gcp.workload_identity_pool.pool_id`, "mywif-pool"))
		Expect(resource).To(MatchJQ(`.attributes.gcp.service_accounts[0].service_account_id`, "mywif-installer"))
	})

	It("Can look up WIF config by id", func() {
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/gcp/wif_configs/wif-config-456"),
				RespondWithJSON(http.StatusOK, `{
				  "id": "wif-config-456",
				  "display_name": "by-id-wif",
				  "organization": {"id": "org-456"},
				  "gcp": {
				    "project_id": "my-gcp-project",
				    "project_number": "123456789",
				    "role_prefix": "byidwif"
				  }
				}`),
			),
		)

		Terraform.Source(`
		  data "osdgoogle_wif_config" "wif" {
		    id = "wif-config-456"
		  }
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("osdgoogle_wif_config", "wif")
		Expect(resource).To(MatchJQ(`.attributes.id`, "wif-config-456"))
		Expect(resource).To(MatchJQ(`.attributes.display_name`, "by-id-wif"))
	})

	It("Fails when WIF config is not found by display_name", func() {
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/gcp/wif_configs"),
				VerifyFormKV("search", "display_name = 'missing-wif'"),
				RespondWithJSON(http.StatusOK, `{"items": []}`),
			),
		)

		Terraform.Source(`
		  data "osdgoogle_wif_config" "wif" {
		    display_name = "missing-wif"
		  }
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).ToNot(BeZero())
		Expect(runOutput.Err).To(ContainSubstring("No WIF config found with display_name"))
	})
})
