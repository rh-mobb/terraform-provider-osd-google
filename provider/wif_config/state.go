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

package wif_config

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// WifConfigState holds the Terraform state for a WIF config.
type WifConfigState struct {
	ID          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	Organization types.String `tfsdk:"organization"`

	GCP *WifGcpState `tfsdk:"gcp"`
}

// WifGcpState holds the GCP-specific WIF configuration.
type WifGcpState struct {
	ProjectID             types.String `tfsdk:"project_id"`
	ProjectNumber         types.String `tfsdk:"project_number"`
	RolePrefix            types.String `tfsdk:"role_prefix"`
	FederatedProjectID    types.String `tfsdk:"federated_project_id"`
	FederatedProjectNumber types.String `tfsdk:"federated_project_number"`
	ImpersonatorEmail     types.String `tfsdk:"impersonator_email"`
}
