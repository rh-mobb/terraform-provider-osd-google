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

package common

import (
	"testing"

	. "github.com/onsi/ginkgo/v2/dsl/core" // nolint
	. "github.com/onsi/gomega"             // nolint

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCommon(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Common Helpers Suite")
}

var _ = Describe("ShouldPatchString", func() {
	It("returns empty and false when both are empty", func() {
		val, ok := ShouldPatchString(types.StringNull(), types.StringNull())
		Expect(val).To(Equal(""))
		Expect(ok).To(BeFalse())
	})
	It("returns plan value and true when state is null and plan is set", func() {
		val, ok := ShouldPatchString(types.StringNull(), types.StringValue("x"))
		Expect(val).To(Equal("x"))
		Expect(ok).To(BeTrue())
	})
	It("returns empty and false when both are same", func() {
		val, ok := ShouldPatchString(types.StringValue("a"), types.StringValue("a"))
		Expect(val).To(Equal(""))
		Expect(ok).To(BeFalse())
	})
	It("returns plan value and true when different", func() {
		val, ok := ShouldPatchString(types.StringValue("a"), types.StringValue("b"))
		Expect(val).To(Equal("b"))
		Expect(ok).To(BeTrue())
	})
})

var _ = Describe("HasValue", func() {
	It("returns false for null", func() {
		Expect(HasValue(types.StringNull())).To(BeFalse())
	})
	It("returns false for empty string", func() {
		Expect(HasValue(types.StringValue(""))).To(BeFalse())
	})
	It("returns true for non-empty string", func() {
		Expect(HasValue(types.StringValue("x"))).To(BeTrue())
	})
})
