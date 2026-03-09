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

var _ = Describe("Cluster waiter resource", func() {
	It("Waits for cluster to be ready and sets ready=true", func() {
		clusterReady := `{
		  "id": "cluster-123",
		  "name": "test-cluster",
		  "state": "ready"
		}`

		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/cluster-123"),
				RespondWithJSON(http.StatusOK, clusterReady),
			),
		)

		Terraform.Source(`
		  resource "osdgoogle_cluster_waiter" "wait" {
		    cluster_id = "cluster-123"
		  }
		`)
		applyOutput := Terraform.Apply()
		Expect(applyOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("osdgoogle_cluster_waiter", "wait")
		Expect(resource).To(MatchJQ(`.attributes.cluster_id`, "cluster-123"))
		Expect(resource).To(MatchJQ(`.attributes.ready`, true))

		destroyOutput := Terraform.Destroy()
		Expect(destroyOutput.ExitCode).To(BeZero())
	})
})
