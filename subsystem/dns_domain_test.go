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

var _ = Describe("DNS domain resource", func() {
	It("Can create, read, and destroy a DNS domain", func() {
		createResp := `{
		  "id": "dns-domain-123",
		  "cluster_arch": "multi"
		}`

		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/dns_domains"),
				RespondWithJSON(http.StatusCreated, createResp),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/dns_domains/dns-domain-123"),
				RespondWithJSON(http.StatusOK, createResp),
			),
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/dns_domains/dns-domain-123"),
				RespondWithJSON(http.StatusNoContent, ""),
			),
		)

		Terraform.Source(`
		  resource "osdgoogle_dns_domain" "dns" {
		    cluster_arch = "multi"
		  }
		`)
		applyOutput := Terraform.Apply()
		Expect(applyOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("osdgoogle_dns_domain", "dns")
		Expect(resource).To(MatchJQ(`.attributes.id`, "dns-domain-123"))
		Expect(resource).To(MatchJQ(`.attributes.cluster_arch`, "multi"))

		destroyOutput := Terraform.Destroy()
		Expect(destroyOutput.ExitCode).To(BeZero())
	})
})
