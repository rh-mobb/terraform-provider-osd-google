variable "ocm_token" {
  type        = string
  sensitive   = true
  default     = ""
  description = "OCM offline token (optional; use OSDGOOGLE_TOKEN env var)"
}

variable "gcp_project_id" {
  type        = string
  description = "GCP service project ID (where cluster resources are created)"
}

variable "vpc_host_project_id" {
  type        = string
  description = "GCP host project ID (shared VPC owner)"
}

variable "vpc_name" {
  type        = string
  description = "Name of the shared VPC network"
}

variable "compute_subnet" {
  type        = string
  description = "Subnet name for worker nodes"
}

variable "control_plane_subnet" {
  type        = string
  description = "Subnet name for control plane nodes"
}

variable "cluster_name" {
  type        = string
  default     = "my-shared-vpc-cluster"
  description = "Name of the cluster (must match terraform/wif_config)"
}

variable "openshift_version" {
  type        = string
  default     = "4.21.3"
  description = "OpenShift version (x.y.z)"
}

variable "machine_pools" {
  type = list(object({
    name                = string
    instance_type       = string
    autoscaling_enabled = optional(bool, true)
    min_replicas        = optional(number)
    max_replicas        = optional(number)
    replicas            = optional(number)
    availability_zones = optional(list(string))
    labels              = optional(map(string), {})
    taints              = optional(list(object({
      key   = string
      value = string
      effect = string
    })), [])
    root_volume_size = optional(number)
    secure_boot      = optional(bool, false)
  }))
  default = []
  description = "Additional machine pools. If autoscaling_enabled = true, set min_replicas and max_replicas. If false, set replicas."

  validation {
    condition = alltrue([
      for mp in var.machine_pools : (
        (mp.autoscaling_enabled && mp.min_replicas != null && mp.max_replicas != null && mp.replicas == null) ||
        (!mp.autoscaling_enabled && mp.replicas != null && mp.min_replicas == null && mp.max_replicas == null)
      )
    ])
    error_message = "For each machine pool: if autoscaling_enabled = true, set min_replicas and max_replicas (replicas must be null). If false, set replicas (min_replicas and max_replicas must be null)."
  }
}
