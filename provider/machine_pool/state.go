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

package machine_pool

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MachinePoolState holds the Terraform state for a machine pool.
type MachinePoolState struct {
	ID               types.String `tfsdk:"id"`
	ClusterID        types.String `tfsdk:"cluster_id"`
	Name             types.String `tfsdk:"name"`
	InstanceType     types.String `tfsdk:"instance_type"`
	Replicas         types.Int64  `tfsdk:"replicas"`
	AvailabilityZones types.List  `tfsdk:"availability_zones"`
	Labels           types.Map    `tfsdk:"labels"`
	Taints           types.List   `tfsdk:"taints"`
	RootVolumeSize   types.Int64  `tfsdk:"root_volume_size"`

	Autoscaling *AutoscalingState `tfsdk:"autoscaling"`
	GCP         *GCPMachinePoolState `tfsdk:"gcp"`
}

// AutoscalingState holds autoscaling config.
type AutoscalingState struct {
	MinReplicas types.Int64 `tfsdk:"min_replicas"`
	MaxReplicas types.Int64 `tfsdk:"max_replicas"`
}

// GCPMachinePoolState holds GCP-specific machine pool options.
type GCPMachinePoolState struct {
	SecureBoot types.Bool `tfsdk:"secure_boot"`
}

// TaintState holds a single taint.
type TaintState struct {
	Key    types.String `tfsdk:"key"`
	Value  types.String `tfsdk:"value"`
	Effect types.String `tfsdk:"effect"`
}
