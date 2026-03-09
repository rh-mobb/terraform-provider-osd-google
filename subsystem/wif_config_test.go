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

var _ = Describe("WIF config resource", func() {
	It("Can create, read, and destroy a WIF config", func() {
		createResp := `{
		  "id": "wif-config-123",
		  "display_name": "my-wif-config",
		  "organization": {"id": "org-123"},
		  "gcp": {
		    "project_id": "my-gcp-project",
		    "project_number": "123456789",
		    "role_prefix": "myprefix"
		  }
		}`

		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/gcp/wif_configs"),
				RespondWithJSON(http.StatusCreated, createResp),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/gcp/wif_configs/wif-config-123"),
				RespondWithJSON(http.StatusOK, createResp),
			),
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/gcp/wif_configs/wif-config-123"),
				RespondWithJSON(http.StatusNoContent, ""),
			),
		)

		Terraform.Source(`
		  resource "osdgoogle_wif_config" "wif" {
		    display_name = "my-wif-config"
		    gcp = {
		      project_id     = "my-gcp-project"
		      project_number = "123456789"
		      role_prefix    = "myprefix"
		    }
		  }
		`)
		applyOutput := Terraform.Apply()
		Expect(applyOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("osdgoogle_wif_config", "wif")
		Expect(resource).To(MatchJQ(`.attributes.id`, "wif-config-123"))
		Expect(resource).To(MatchJQ(`.attributes.display_name`, "my-wif-config"))

		destroyOutput := Terraform.Destroy()
		Expect(destroyOutput.ExitCode).To(BeZero())
	})

	It("Can create a WIF config with openshift_version to scope IAM resources", func() {
		createResp := `{
		  "id": "wif-config-456",
		  "display_name": "version-scoped-wif",
		  "organization": {"id": "org-456"},
		  "gcp": {
		    "project_id": "my-gcp-project",
		    "project_number": "123456789",
		    "role_prefix": "osd"
		  }
		}`

		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/gcp/wif_configs"),
				RespondWithJSON(http.StatusCreated, createResp),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/gcp/wif_configs/wif-config-456"),
				RespondWithJSON(http.StatusOK, createResp),
			),
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/gcp/wif_configs/wif-config-456"),
				RespondWithJSON(http.StatusNoContent, ""),
			),
		)

		Terraform.Source(`
		  resource "osdgoogle_wif_config" "wif" {
		    display_name      = "version-scoped-wif"
		    openshift_version = "4.21"
		    gcp = {
		      project_id     = "my-gcp-project"
		      project_number = "123456789"
		      role_prefix    = "osd"
		    }
		  }
		`)
		applyOutput := Terraform.Apply()
		Expect(applyOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("osdgoogle_wif_config", "wif")
		Expect(resource).To(MatchJQ(`.attributes.id`, "wif-config-456"))
		Expect(resource).To(MatchJQ(`.attributes.openshift_version`, "4.21"))

		destroyOutput := Terraform.Destroy()
		Expect(destroyOutput.ExitCode).To(BeZero())
	})
})
