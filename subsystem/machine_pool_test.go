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

var _ = Describe("Machine pool resource", func() {
	It("Can create and destroy a machine pool", func() {
		TestServer.AppendHandlers(
			CombineHandlers(
				VerifyRequest(http.MethodPost, "/api/clusters_mgmt/v1/clusters/cluster-123/machine_pools"),
				RespondWithJSON(http.StatusCreated, `{
				  "id": "worker",
				  "instance_type": "custom-4-16384",
				  "replicas": 3
				}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodGet, "/api/clusters_mgmt/v1/clusters/cluster-123/machine_pools/worker"),
				RespondWithJSON(http.StatusOK, `{
				  "id": "worker",
				  "instance_type": "custom-4-16384",
				  "replicas": 3
				}`),
			),
			CombineHandlers(
				VerifyRequest(http.MethodDelete, "/api/clusters_mgmt/v1/clusters/cluster-123/machine_pools/worker"),
				RespondWithJSON(http.StatusNoContent, ""),
			),
		)

		Terraform.Source(`
		  resource "osdgoogle_machine_pool" "pool" {
		    cluster_id     = "cluster-123"
		    name           = "worker"
		    instance_type  = "custom-4-16384"
		    replicas       = 3
		  }
		`)
		applyOutput := Terraform.Apply()
		Expect(applyOutput.ExitCode).To(BeZero())

		resource := Terraform.Resource("osdgoogle_machine_pool", "pool")
		Expect(resource).To(MatchJQ(`.attributes.id`, "worker"))

		destroyOutput := Terraform.Destroy()
		Expect(destroyOutput.ExitCode).To(BeZero())
	})
})
