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

var _ = Describe("Machine types data source", func() {
	It("Can list machine types for a region", func() {
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/gcp_inquiries/machine_types"),
				RespondWithJSON(http.StatusOK, `{
				  "items": [
				    {"id": "custom-4-16384", "name": "custom-4-16384"},
				    {"id": "n2-standard-4", "name": "n2-standard-4"}
				  ]
				}`),
			),
		)

		Terraform.Source(`
		  data "osdgoogle_machine_types" "my_types" {
		    region         = "us-central1"
		    gcp_project_id = "my-gcp-project"
		  }
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("osdgoogle_machine_types", "my_types")
		Expect(resource).To(MatchJQ(`.attributes.items[0].id`, "custom-4-16384"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].id`, "n2-standard-4"))
	})
})
