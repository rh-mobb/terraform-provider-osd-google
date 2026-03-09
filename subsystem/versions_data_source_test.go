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

var _ = Describe("Versions data source", func() {
	It("Can list versions", func() {
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/versions"),
				VerifyFormKV("search", "enabled = 't'"),
				RespondWithJSON(http.StatusOK, `{
				  "page": 1,
				  "size": 2,
				  "total": 2,
				  "items": [
				    {
				      "id": "openshift-v4.16.1",
				      "raw_id": "4.16.1"
				    },
				    {
				      "id": "openshift-v4.16.2",
				      "raw_id": "4.16.2"
				    }
				  ]
				}`),
			),
		)

		Terraform.Source(`
		  data "osdgoogle_versions" "my_versions" {
		  }
		`)
		runOutput := Terraform.Apply()
		Expect(runOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("osdgoogle_versions", "my_versions")
		Expect(resource).To(MatchJQ(`.attributes.items[0].id`, "openshift-v4.16.1"))
		Expect(resource).To(MatchJQ(`.attributes.items[0].name`, "4.16.1"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].id`, "openshift-v4.16.2"))
		Expect(resource).To(MatchJQ(`.attributes.items[1].name`, "4.16.2"))
	})
})
