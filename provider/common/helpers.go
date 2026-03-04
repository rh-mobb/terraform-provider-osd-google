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
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ShouldPatchString checks if the change between state and plan requires a patch.
func ShouldPatchString(state, plan types.String) (value string, ok bool) {
	if plan.IsUnknown() || plan.IsNull() {
		return
	}
	if state.IsUnknown() || state.IsNull() {
		return plan.ValueString(), true
	}
	if plan.ValueString() != state.ValueString() {
		return plan.ValueString(), true
	}
	return
}

// ShouldPatchBool checks if the change between state and plan requires a patch.
func ShouldPatchBool(state, plan types.Bool) (value bool, ok bool) {
	if plan.IsUnknown() || plan.IsNull() {
		return
	}
	if state.IsUnknown() || state.IsNull() {
		return plan.ValueBool(), true
	}
	if plan.ValueBool() != state.ValueBool() {
		return plan.ValueBool(), true
	}
	return
}

// ShouldPatchInt64 checks if the change between state and plan requires a patch.
func ShouldPatchInt64(state, plan types.Int64) (value int64, ok bool) {
	if plan.IsUnknown() || plan.IsNull() {
		return
	}
	if state.IsUnknown() || state.IsNull() {
		return plan.ValueInt64(), true
	}
	if plan.ValueInt64() != state.ValueInt64() {
		return plan.ValueInt64(), true
	}
	return
}

// ShouldPatchMap checks if the change between state and plan requires a patch.
func ShouldPatchMap(state, plan types.Map) (types.Map, bool) {
	return plan, !reflect.DeepEqual(state.Elements(), plan.Elements())
}

// HasValue returns true if the string attribute has a non-null, non-empty value.
func HasValue(attr types.String) bool {
	return !attr.IsNull() && !attr.IsUnknown() && attr.ValueString() != ""
}

// OptionalString returns a pointer to the string value or nil if null/empty.
func OptionalString(attr types.String) *string {
	if !HasValue(attr) {
		return nil
	}
	s := attr.ValueString()
	return &s
}

// OptionalInt64 returns a pointer to the int64 value or nil if null.
func OptionalInt64(attr types.Int64) *int64 {
	if attr.IsNull() || attr.IsUnknown() {
		return nil
	}
	v := attr.ValueInt64()
	return &v
}

// StringListToArray converts a Terraform list of strings to a Go string slice.
func StringListToArray(list types.List) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	elements := list.Elements()
	result := make([]string, 0, len(elements))
	for _, elem := range elements {
		if str, ok := elem.(types.String); ok && !str.IsNull() {
			result = append(result, str.ValueString())
		}
	}
	return result
}
