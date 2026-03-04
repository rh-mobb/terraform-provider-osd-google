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

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestShouldPatchString(t *testing.T) {
	tests := []struct {
		name    string
		state   types.String
		plan    types.String
		wantVal string
		wantOk  bool
	}{
		{"both empty", types.StringNull(), types.StringNull(), "", false},
		{"state null plan set", types.StringNull(), types.StringValue("x"), "x", true},
		{"both same", types.StringValue("a"), types.StringValue("a"), "", false},
		{"different", types.StringValue("a"), types.StringValue("b"), "b", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := ShouldPatchString(tt.state, tt.plan)
			if val != tt.wantVal || ok != tt.wantOk {
				t.Errorf("ShouldPatchString() = (%q, %v), want (%q, %v)", val, ok, tt.wantVal, tt.wantOk)
			}
		})
	}
}

func TestHasValue(t *testing.T) {
	if HasValue(types.StringNull()) {
		t.Error("HasValue(null) should be false")
	}
	if HasValue(types.StringValue("")) {
		t.Error("HasValue(empty) should be false")
	}
	if !HasValue(types.StringValue("x")) {
		t.Error("HasValue(x) should be true")
	}
}
