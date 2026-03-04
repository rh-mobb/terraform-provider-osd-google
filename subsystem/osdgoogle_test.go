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
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint
	"github.com/onsi/gomega/format"
	. "github.com/openshift-online/ocm-sdk-go/testing" // nolint
	. "github.com/redhat/terraform-provider-osd-google/subsystem/framework"
)

func TestSubsystem(t *testing.T) {
	RegisterFailHandler(Fail)
	TestingT = t
	RunSpecs(t, "OSD Google Subsystem")
}

var _ = BeforeEach(func() {
	format.MaxLength = 0
	var ca string
	TestServer, ca = MakeTCPTLSServer()
	token := MakeTokenString("Bearer", 10*time.Minute)

	Terraform = NewTerraformRunner().
		URL(TestServer.URL()).
		CA(ca).
		Token(token).
		Build()
})

var _ = AfterEach(func() {
	TestServer.Close()
	Terraform.Close()
})
