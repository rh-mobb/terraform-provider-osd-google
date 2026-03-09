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

package cluster_admin

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ClusterAdminState holds the Terraform state for a cluster admin resource.
type ClusterAdminState struct {
	ID         types.String `tfsdk:"id"`
	ClusterID   types.String `tfsdk:"cluster_id"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}
